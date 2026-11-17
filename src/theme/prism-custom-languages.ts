/**
 * Small Prism grammars for languages the docs use that have no prismjs
 * component: kdl (niri/mangowm configs), conf (sway/hyprland-style configs),
 * pam and cron. Registered from prism-include-languages.ts.
 *
 * Token names map onto the dankPurple themes in docusaurus.config.ts:
 *   property/attr-name = blue (keys)    string/attr-value = green (values)
 *   tag = rose (node names)             keyword = purple
 *   class-name/builtin = teal           function = violet
 *   number/boolean = amber              variable = cyan
 *
 * Only type imports here, so the file can also be loaded by plain node for tests.
 */
import type * as PrismNamespace from 'prismjs';

type PrismType = typeof PrismNamespace;

export function registerCustomLanguages(Prism: PrismType): void {
  const {languages} = Prism;

  // ---- KDL --------------------------------------------------------------
  languages.kdl = {
    comment: [
      {pattern: /\/\*[\s\S]*?\*\//, greedy: true},
      {pattern: /\/\/.*/, greedy: true},
    ],
    string: [
      {pattern: /r(#*)"[\s\S]*?"\1/, greedy: true},
      {pattern: /"""[\s\S]*?"""/, greedy: true},
      {pattern: /"(?:\\[\s\S]|[^\\"])*"/, greedy: true},
    ],
    // `/-` comments out the next node, argument or property.
    slashdash: {pattern: /\/-/, alias: 'important'},
    // First identifier of a node: line start or right after `{`, `}` or `;`.
    tag: {
      pattern: /(^[ \t]*(?:\/-[ \t]*)?|[{};][ \t]*(?:\/-[ \t]*)?)(?:\([^()\s]*\))?[^\s{};()="\\/#][^\s{};()="\\]*/m,
      lookbehind: true,
      greedy: true,
      inside: {
        'type-annotation': {pattern: /^\([^()\s]*\)/, alias: 'class-name'},
      },
    },
    'type-annotation': {
      pattern: /\([^()\s]+\)/,
      alias: 'class-name',
      inside: {punctuation: /[()]/},
    },
    property: {
      pattern: /(^|[\s{;])[^\s{};()="\\/#][^\s{};()="\\]*(?==)/,
      lookbehind: true,
    },
    boolean: {
      pattern: /(^|[\s=(])#?(?:true|false)(?=[\s;})]|$)/m,
      lookbehind: true,
    },
    null: {
      pattern: /(^|[\s=(])#?null(?=[\s;})]|$)/m,
      lookbehind: true,
      alias: 'keyword',
    },
    number: {
      pattern: /(^|[\s=(])[+-]?(?:0x[\da-fA-F_]+|0o[0-7_]+|0b[01_]+|\d[\d_]*(?:\.\d[\d_]*)?(?:[eE][+-]?\d+)?)(?=[\s;})]|$)/m,
      lookbehind: true,
    },
    operator: /=/,
    punctuation: /[{}();]/,
  };

  // ---- conf (sway / i3 / hyprland / mangowm style) ------------------------
  languages.conf = {
    // `#` starts a comment unless it is a hex colour like #ff00aa.
    comment: {
      pattern: /(^|[ \t])#(?![0-9a-fA-F]{3,8}\b).*/m,
      lookbehind: true,
      greedy: true,
    },
    string: {pattern: /"(?:\\.|[^\\"\r\n])*"|'[^'\r\n]*'/, greedy: true},
    // `key = value` and `key: value`
    property: {
      pattern: /^[ \t]*[\w.$:-]+(?=[ \t]*=)/m,
      greedy: true,
    },
    // `name {` blocks
    'class-name': {
      pattern: /^[ \t]*[\w.$:-]+(?=[ \t]*\{)/m,
      greedy: true,
    },
    // sway/i3 style: first word of a `command args...` line
    keyword: {
      pattern: /^[ \t]*[A-Za-z_][\w-]*(?=[ \t]+\S)(?![^\n=]*=)/m,
      greedy: true,
    },
    variable: /\$\{?[\w.-]+\}?/,
    color: {pattern: /#[0-9a-fA-F]{3,8}\b/, alias: 'constant'},
    boolean: /\b(?:true|false|yes|no|on|off)\b/i,
    number: /\b\d+(?:\.\d+)?\b/,
    operator: /[=,+]/,
    punctuation: /[{}()[\]]/,
  };

  // ---- PAM ----------------------------------------------------------------
  const pamBracket = {
    pattern: /\[[^\]\n]*\]/,
    alias: 'builtin',
    inside: {
      property: /[\w-]+(?==)/,
      number: /\b\d+\b/,
      'attr-value': {pattern: /(=)[^\s\]]+/, lookbehind: true},
      punctuation: /[[\]=]/,
    },
  };
  languages.pam = {
    comment: {pattern: /#.*/, greedy: true},
    line: {
      pattern: /^[ \t]*(?:-?(?:auth|account|password|session)|@include)\b.*/m,
      inside: {
        type: {
          pattern: /^[ \t]*(?:-?(?:auth|account|password|session)|@include)\b/,
          alias: 'keyword',
        },
        control: {
          pattern: /(\s)(?:required|requisite|sufficient|optional|include|substack|binding)\b/,
          lookbehind: true,
          alias: 'builtin',
        },
        bracket: pamBracket,
        module: {
          pattern: /(\s)(?:[\w./-]+\.so|pam_\w+)\b/,
          lookbehind: true,
          alias: 'function',
        },
        property: {pattern: /(\s)[\w-]+(?==)/, lookbehind: true},
        number: {pattern: /(=)\d+\b/, lookbehind: true},
        'attr-value': {pattern: /(=)\S+/, lookbehind: true},
        operator: /=/,
      },
    },
  };

  // ---- cron ---------------------------------------------------------------
  const cronFields = {
    constant: {
      pattern: /\b(?:sun|mon|tue|wed|thu|fri|sat|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\b/i,
    },
    number: /\b\d+\b/,
    operator: /[*/,\-?]/,
  };
  languages.cron = {
    comment: {pattern: /#.*/, greedy: true},
    // MAILTO=me@host, PATH=/usr/bin
    env: {
      pattern: /^[ \t]*[A-Za-z_]\w*[ \t]*=.*/m,
      inside: {
        variable: /^[ \t]*\w+/,
        operator: /=/,
        string: /\S.*/,
      },
    },
    // @reboot, @daily, ... command
    'special-job': {
      pattern: /^[ \t]*@(?:reboot|yearly|annually|monthly|weekly|daily|midnight|hourly)\b.*/im,
      inside: {
        special: {pattern: /^[ \t]*@\w+/, alias: 'keyword'},
        command: {pattern: /\S.*/, inside: languages.bash ?? {}},
      },
    },
    // m h dom mon dow command
    job: {
      pattern: /^[ \t]*[\d*/,\-?LW#a-z]+(?:[ \t]+[\d*/,\-?LW#a-z]+){4}[ \t]+\S.*/im,
      inside: {
        schedule: {
          pattern: /^[ \t]*\S+(?:[ \t]+\S+){4}/,
          inside: cronFields,
        },
        command: {pattern: /\S.*/, inside: languages.bash ?? {}},
      },
    },
  };
}
