{
  local root = self,

  templates: {
    package: |||
      # %(name)s

      %(content)s
    |||,

    indexPage: |||
      # %(prefix)s%(name)s

      %(index)s
    |||,

    index: |||
      ## Index

      %s
    |||,

    sectionTitle: '%(abbr)s %(prefix)s%(name)s',

    sectionLink: '* [`%(abbr)s %(linkName)s`](#%(link)s)',

    value: '* `%(prefix)s%(name)s` (`%(type)s`): `"%(value)s"` - %(help)s',

    section: |||
      %(headerDepth)s %(title)s

      %(content)s
    |||,
  },

  joinPathPrefixes(prefixes, sep='/')::
    std.join(sep, prefixes)
    + (if std.length(prefixes) > 0
       then sep
       else ''),

  joinPrefixes(prefixes, sep='.')::
    std.join(sep, prefixes)
    + (if std.length(prefixes) > 0
       then sep
       else ''),

  renderSectionTitle(section, prefixes)::
    root.templates.sectionTitle % {
      name: section.name,
      abbr: section.type.abbr,
      prefix: root.joinPrefixes(prefixes),
    },

  renderValues(values, prefixes=[])::
    if std.length(values) > 0
    then
      std.join('\n', [
        root.templates.value
        % value {
          prefix: root.joinPrefixes(prefixes),
        }
        for value in values
      ]) + '\n'
    else '',

  renderSections(sections, depth=0, prefixes=[])::
    if std.length(sections) > 0
    then
      std.join('\n', [
        root.templates.section
        % {
          headerDepth: std.join('', [
            '#'
            for d in std.range(0, depth + 2)
          ]),
          title: root.renderSectionTitle(
            section,
            prefixes,
          ),
          content: section.content,
        }
        + root.renderValues(
          section.values,
          prefixes + [section.name]
        )
        + root.renderSections(
          section.subSections,
          depth + 1,
          prefixes + [section.name]
        )
        for section in sections
      ])
    else '',

  renderPackage(package, path='')::
    (root.templates.package % package)
    + (
      if std.length(package.subPackages) > 0
      then
        '## Subpackages\n\n'
        + std.join('\n', [
          '* [%(name)s](%(path)s)' % {
            name: sub.name,
            path: path + sub.name
                  + (if std.length(sub.subPackages) > 0
                     then '/index.md'
                     else '.md'),
          }
          for sub in package.subPackages
        ]) + '\n\n'
      else ''
    )
    + (if std.length(package.sections) > 0
       then (root.templates.index % root.index(package.sections))
       else '')
    + (if std.length(package.values) > 0
          || std.length(package.sections) > 0
       then
         '\n## Fields\n\n'
         + root.renderValues(package.values)
         + root.renderSections(package.sections)
       else ''),

  index(sections, depth=0, prefixes=[])::
    std.join('\n', [
      std.join('', [
        ' '
        for d in std.range(0, (depth * 2) - 1)
      ])
      + (root.templates.sectionLink % {
           abbr: section.type.abbr,
           linkName: section.linkName,
           link:
             std.asciiLower(
               std.strReplace(
                 std.strReplace(root.renderSectionTitle(section, prefixes), '.', '')
                 , ' ', '-'
               )
             ),
         })
      + (
        if std.length(section.subSections) > 0
        then '\n' + root.index(
          section.subSections,
          depth + 1,
          prefixes + [section.name]
        )
        else ''
      )
      for section in sections
    ]),

  sections: {
    base: {
      subSections: [],
      values: [],
    },
    object(key, doc, obj, depth):: self.base {
      name: std.strReplace(key, '#', ''),

      local processed = root.prepare(obj, depth=depth + 1),

      subSections: processed.sections,

      values: processed.values,

      type: { full: 'object', abbr: 'obj' },

      abbr: self.type.abbr,

      doc:
        if self.type.full in doc
        then doc[self.type.full]
        else { help: '' },

      help: self.doc.help,

      linkName: self.name,

      content:
        if self.help != ''
        then self.help + '\n'
        else '',
    },

    'function'(key, doc):: self.base {
      name: std.strReplace(key, '#', ''),

      type: { full: 'function', abbr: 'fn' },

      abbr: self.type.abbr,

      doc: doc[self.type.full],

      help: self.doc.help,

      args: std.join(', ', [
        if arg.default != null
        then std.join('=', [
          arg.name,
          std.manifestJsonEx(arg.default, '', ''),
        ])
        else arg.name
        for arg in self.doc.args
      ]),

      enums: std.join('', [
        if arg.enums != null
        then '\n\nAccepted values for `%s` are ' % arg.name
             + std.join(', ', [
               std.manifestJsonEx(item, '', '')
               for item in arg.enums
             ])
        else ''
        for arg in self.doc.args
      ]),

      linkName: '%(name)s(%(args)s)' % self,

      content:
        (|||
           ```ts
           %(name)s(%(args)s)
           ```

         ||| % self)
        + '%(help)s' % self
        + '%(enums)s' % self,
      // odd concatenation to prevent unintential newline changes

    },

    value(key, doc, obj):: self.base {
      name: std.strReplace(key, '#', ''),
      type: doc.value.type,
      help: doc.value.help,
      value: obj,
    },

    package(doc, root):: {
      name: doc.name,
      content:
        |||
          %(help)s
        ||| % doc
        + (if 'installTemplate' in doc
           then |||

             ## Install

             ```
             %(install)s
             ```
           ||| % doc.installTemplate % doc
           else '')
        + (if 'usageTemplate' in doc
           then |||

             ## Usage

             ```jsonnet
             %(usage)s
             ```
           ||| % doc.usageTemplate % doc
           else ''),
    },
  },

  prepare(obj, depth=0)::
    std.foldl(
      function(acc, key)
        acc +
        // Package definition
        if key == '#'
        then root.sections.package(
          obj[key],
          (depth == 0)
        )


        // Field definition
        else if std.startsWith(key, '#')
        then (
          local realKey = key[1:];

          if !std.isObject(obj[key])
          then
            std.trace(
              'INFO: docstring "%s" cannot be parsed, ignored while rendering.' % key,
              {}
            )

          else if 'value' in obj[key]
          then {
            values+: [root.sections.value(
              key,
              obj[key],
              obj[realKey]
            )],
          }
          else if 'function' in obj[key]
          then {
            functionSections+: [root.sections['function'](
              key,
              obj[key],
            )],
          }
          else if 'object' in obj[key]
          then {
            objectSections+: [root.sections.object(
              key,
              obj[key],
              obj[realKey],
              depth
            )],
          }
          else
            std.trace(
              'INFO: docstring "%s" cannot be parsed, ignored while rendering.' % key,
              {}
            )
        )

        // subPackage definition
        else if std.isObject(obj[key]) && '#' in obj[key]
        then {
          subPackages+: [root.prepare(obj[key])],
        }

        // undocumented object
        else if std.isObject(obj[key]) && !('#' + key in obj)
        then (
          local section = root.sections.object(
            key,
            {},
            obj[key],
            depth
          );
          // only add if has documented subSections or values
          if std.length(section.subSections) > 0
             || std.length(section.values) > 0
          then { objectSections+: [section] }
          else {}
        )

        else {},
      std.objectFieldsAll(obj),
      {
        functionSections: [],
        objectSections: [],

        sections:
          self.functionSections
          + self.objectSections,
        subPackages: [],
        values: [],
      }
    ),

  renderIndexPage(package, prefixes)::
    root.templates.indexPage % {
      name: package.name,
      prefix: root.joinPrefixes(prefixes),
      index: std.join('\n', [
        '* [%(name)s](%(name)s.md)' % sub
        for sub in package.subPackages
      ]),
    },

  renderFiles(package, prefixes=[]):
    local path = root.joinPathPrefixes(prefixes);
    (
      if std.length(prefixes) == 0
      then {
        [path + 'README.md']: root.renderPackage(package, package.name + '/'),
      }
      else if std.length(package.subPackages) > 0
      then {
        [path + package.name + '/index.md']: root.renderPackage(package),
      }
      else {
        [path + package.name + '.md']: root.renderPackage(package, package.name + '/'),
      }
    )
    + std.foldl(
      function(acc, sub)
        acc + sub,
      [
        root.renderFiles(
          sub,
          prefixes=prefixes + [package.name]
        )
        for sub in package.subPackages
      ],
      {}
    ),

  render(obj):
    self.renderFiles(self.prepare(obj)),
}
