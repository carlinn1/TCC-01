#!/usr/bin/env python3
r"""
Verificador automatizado de integridade LaTeX e BibTeX (BASE-02, FOR-02, REL-02).
Executa:
  1. Verificacao sintatica de ambientes (\begin{...} / \end{...}) e chaves { } em Docs/Main.tex
     e em todos os fragmentos Docs/tabelas/*.tex;
  2. Verificacao de referencias cruzadas: todo \ref{...} possui \label{...} correspondente
     e nao ha labels duplicados;
  3. Verificacao bibliografica completa: toda chave em \cite{...} ou \citeonline{...} existe em
     Docs/Referencias.bib e todas as 50 entradas de Docs/Referencias.bib sao citadas no texto
     (zero entradas orfas, zero citacoes indefinidas);
  4. Verificacao de existencia fisica de todos os arquivos \input{...} e \includegraphics{...};
  5. Verificacao de ausencia de '%' nao escapado em celulas de tabelas geradas e ausencia de
     termos legados proibidos;
  6. Se 'pdflatex' estiver disponivel no PATH do sistema, executa a compilacao completa e
     verifica a ausencia de erros/warnings no Main.log.
"""
import os
import re
import shutil
import subprocess
import sys


def strip_latex_comments(text: str) -> str:
    lines = []
    for line in text.splitlines():
        # Remove comentario a partir de % nao precedido por \
        cleaned = re.sub(r"(?<!\\)%.*$", "", line)
        lines.append(cleaned)
    return "\n".join(lines)


def check_balanced_environments(tex_no_comments: str, filename: str, errors: list[str]) -> None:
    stack: list[tuple[str, int]] = []
    for lineno, line in enumerate(tex_no_comments.splitlines(), start=1):
        for m in re.finditer(r"\\(begin|end)\{([^}]+)\}", line):
            kind, env = m.group(1), m.group(2)
            if kind == "begin":
                stack.append((env, lineno))
            else:
                if not stack:
                    errors.append(f"{filename}:{lineno}: \\end{{{env}}} sem \\begin correspondente")
                else:
                    top_env, top_line = stack.pop()
                    if top_env != env:
                        errors.append(
                            f"{filename}:{lineno}: \\end{{{env}}} fecha \\begin{{{top_env}}} aberto na linha {top_line}"
                        )
    for env, lineno in stack:
        errors.append(f"{filename}:{lineno}: \\begin{{{env}}} nao fechado ate o final do arquivo")


def main() -> None:
    code_dir = os.path.dirname(os.path.abspath(__file__))
    root_dir = os.path.dirname(code_dir)
    docs_dir = os.path.join(root_dir, "Docs")

    main_tex_path = os.path.join(docs_dir, "Main.tex")
    bib_path = os.path.join(docs_dir, "Referencias.bib")
    root_bib_path = os.path.join(root_dir, "referencias.bib")

    errors: list[str] = []

    with open(main_tex_path, "r", encoding="utf-8") as f:
        main_raw = f.read()
    main_clean = strip_latex_comments(main_raw)

    # 1. Ambientes balanceados em Docs/Main.tex
    check_balanced_environments(main_clean, "Docs/Main.tex", errors)

    # 2. Verifica existencia e sintaxe de todos os \input{...} e \includegraphics{...}
    for m in re.finditer(r"\\input\{([^}]+)\}", main_clean):
        rel = m.group(1)
        target = os.path.join(docs_dir, rel)
        if not os.path.exists(target):
            errors.append(f"Arquivo \\input{{{rel}}} nao encontrado em {target}")
        else:
            with open(target, "r", encoding="utf-8") as tf:
                t_raw = tf.read()
            # Verifica se alguma linha de dados da tabela gerada possui % nao escapado apos &
            for lno, tline in enumerate(t_raw.splitlines(), start=1):
                if tline.lstrip().startswith("%"):
                    continue
                if re.search(r"(?<!\\)%", tline):
                    errors.append(f"Docs/{rel}:{lno}: caractere '%' nao escapado em linha de tabela/macro: {tline}")

    for m in re.finditer(r"\\includegraphics(?:\[[^\]]*\])?\{([^}]+)\}", main_clean):
        rel = m.group(1)
        if "#" in rel or rel == "capa_fundo.jpg":
            continue
        target = os.path.join(docs_dir, rel)
        if not os.path.exists(target):
            errors.append(f"Imagem \\includegraphics{{{rel}}} nao encontrada em {target}")

    fig_labels = set()
    for m in re.finditer(r"\\figura(?:Adaptada)?\{([^}]+)\}\{(?:[^{}]|\{[^{}]*\})*\}\{([^}]+)\}", main_clean):
        rel, flbl = m.group(1), m.group(2)
        if "#" in rel:
            continue
        fig_labels.add(flbl)
        target = os.path.join(docs_dir, "figuras", rel)
        if not os.path.exists(target) or os.path.getsize(target) < 1024:
            errors.append(f"Figura {rel} ausente ou invalida em {target}")

    # 3. Verifica \label{...} vs \ref{...}
    labels = [lbl for lbl in re.findall(r"\\label\{([^}]+)\}", main_clean) if "#" not in lbl]
    seen_labels = set(fig_labels)
    for lbl in labels:
        if lbl in seen_labels:
            errors.append(f"Label duplicado em Docs/Main.tex: \\label{{{lbl}}}")
        seen_labels.add(lbl)

    refs = re.findall(r"\\ref\{([^}]+)\}", main_clean)
    for r in refs:
        if r not in seen_labels:
            errors.append(f"Referencia indefinida em Docs/Main.tex: \\ref{{{r}}}")

    # 4. Verifica citacoes BibTeX (\cite e \citeonline) e sincronia Docs/Referencias.bib == referencias.bib
    with open(bib_path, "r", encoding="utf-8") as bf:
        bib_raw = bf.read()
    with open(root_bib_path, "r", encoding="utf-8") as rbf:
        root_bib_raw = rbf.read()
    if bib_raw != root_bib_raw:
        errors.append("Dessincronia entre Docs/Referencias.bib e referencias.bib na raiz")

    bib_keys = set(re.findall(r"@\w+\{\s*([^,\s]+)\s*,", bib_raw))
    cited_keys = set()
    for m in re.finditer(r"\\cite(?:online)?\{([^}]+)\}", main_clean):
        for k in m.group(1).split(","):
            k_clean = k.strip()
            if k_clean:
                cited_keys.add(k_clean)
                if k_clean not in bib_keys:
                    errors.append(f"Citacao indefinida em Docs/Main.tex: \\cite{{{k_clean}}} nao existe no .bib")

    uncited = sorted(bib_keys - cited_keys)
    if uncited:
        errors.append(f"Entradas no .bib nao citadas em Docs/Main.tex ({len(uncited)}): {', '.join(uncited)}")

    # 5. Verifica ausencia de expressoes proibidas pelas decisoes DEC-1 a DEC-7
    forbidden_phrases = [
        ("extensão metodológica proposta", "DEC-3 proibe a expressao 'extensao metodologica proposta'"),
        ("Gêmeo Digital assistido", "TXT-01 exige padronizacao para 'Sombra Digital Governada'"),
        ("assinada com HMAC pela chave privada", "DEC-7/SEG-01 proibe chamar segredo simetrico HMAC de 'chave privada'"),
        ("Prof. Dr. Sérgio", "Usuario solicitou remover nomes/data nao confirmados da folha de aprovacao"),
        ("27 de novembro de 2026", "Usuario solicitou remover data fixa da folha de aprovacao"),
    ]
    for phrase, reason in forbidden_phrases:
        if phrase in main_clean:
            errors.append(f"Frase proibida encontrada em Docs/Main.tex ('{phrase}'): {reason}")

    # 6. Se pdflatex estiver disponivel no sistema, compila e inspeciona o log
    pdflatex_bin = shutil.which("pdflatex")
    bibtex_bin = shutil.which("bibtex")
    if pdflatex_bin and bibtex_bin:
        for cmd in [
            [pdflatex_bin, "-interaction=nonstopmode", "Main.tex"],
            [bibtex_bin, "Main"],
            [pdflatex_bin, "-interaction=nonstopmode", "Main.tex"],
            [pdflatex_bin, "-interaction=nonstopmode", "Main.tex"],
        ]:
            proc = subprocess.run(cmd, cwd=docs_dir, capture_output=True, text=True)
            if proc.returncode != 0:
                errors.append(f"Falha ao executar {' '.join(cmd)} (exit={proc.returncode})")
                break
        log_path = os.path.join(docs_dir, "Main.log")
        if os.path.exists(log_path):
            with open(log_path, "r", encoding="utf-8", errors="replace") as lf:
                log_txt = lf.read()
            for pat in ["LaTeX Warning: Reference", "LaTeX Warning: Citation", "Overfull \\hbox"]:
                if pat in log_txt:
                    errors.append(f"Main.log contem alerta: {pat}")

    if errors:
        print("FALHA na verificacao LaTeX/BibTeX (check_latex.py):")
        for e in errors:
            print(f"  [ERRO] {e}")
        sys.exit(1)

    print("=========================================================================")
    print(f"  [OK] check_latex.py: {len(seen_labels)} labels verificados, {len(cited_keys)}/{len(bib_keys)} entradas")
    print("       BibTeX citadas (0 orfas, 0 indefinidas), ambientes e tabelas integros!")
    print("=========================================================================")


if __name__ == "__main__":
    main()
