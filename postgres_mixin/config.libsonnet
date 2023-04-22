{
  _config+:: {
    dbNameFilter: 'datname!~"template.*"',
    postgresExporterSelector: 'job="integrations/postgres_exporter"',
    groupLabels: if self.enableMultiCluster then ['job', 'cluster'] else ['job'],
    instanceLabels: ['instance', 'server'],
    enableMultiCluster: false,
  },
}
