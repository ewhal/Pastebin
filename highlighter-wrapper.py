#!/usr/bin/env python

try:
    import pygments
except ImportError:
    print(" Please install python pygments module")

from pygments import highlight
from pygments.lexers import get_lexer_by_name, guess_lexer
from pygments.formatters import HtmlFormatter
import sys

def render(code, lang, theme):

    guess = ""
    lang_org = lang

    try:
        lexer = get_lexer_by_name(lang)
    except:
        try:
            guess = 1
            lexer = guess_lexer(code)
            lang  = lexer.aliases[0]
        except:
            if lang == "autodetect":
                out = "Could not autodetect language (returning plain text).\n"
            else:
                out = "Given language was not found :: '"+lang+"' (returning plain text).\n"

            lexer = get_lexer_by_name("text")
            html_format = HtmlFormatter(style=theme, noclasses="true", linenos="true", encoding="utf-8")
            return highlight(code, lexer, html_format),out

    if guess:
        out = "Lexer guessed :: "+lang
        if lang != lang_org and lang_org != "autodetect":
            out += " (although given language was "+lang_org+") "
    else:
        out = "Successfully used lexer for given language :: "+lang

    try:
        html_format = HtmlFormatter(style=theme, noclasses="true", linenos="true", encoding="utf-8")
    except:
        html_format = HtmlFormatter(noclasses="true", linenos="true", encoding="utf-8")

    return highlight(code, lexer, html_format),out



def usage(err=0):
    print("\n Description, \n")
    print("    - This is a small wrapper for the pygments html-formatter.")
    print("      It will read data on stdin and simply print it on stdout")

    print("\n Usage, \n")
    print("    - %s [lang] [style] < FILE" % sys.argv[0])
    print("    - %s getlexers" % sys.argv[0])
    print("    - %s getstyles" % sys.argv[0])

    print("\n Where, \n")
    print("    - lang is the language of your code")
    print("    - style is the 'theme' for the formatter")
    print("    - getlexers will print available lexers (displayname;lexer-name)")
    print("    - getstyles will print available styles \n")

    sys.exit(err)

def get_styles():
    item = pygments.styles.get_all_styles()
    for items in item:
        print(items)
    sys.exit(0)

def get_lexers():
    item = pygments.lexers.get_all_lexers()
    for items in item:
        print(items[0]+";"+items[1][0])
    sys.exit(0)


# " Main "

code = ""

if len(sys.argv) >= 2:
  for arg in sys.argv:
      if arg == '--help' or arg == '-h':
          usage()
      if arg == 'getlexers':
          get_lexers()
      if arg == 'getstyles':
          get_styles()

if len(sys.argv) == 3:
    lang  = sys.argv[1]
    theme = sys.argv[2]
else:
    usage(1);

if not sys.stdin.isatty():
    for line in sys.stdin:
        code += line

    out, stderr = render(code, lang, theme)
    print(out)
    sys.stderr.write(stderr)
else:
    print("err : No data on stdin.")
    sys.exit(1)
