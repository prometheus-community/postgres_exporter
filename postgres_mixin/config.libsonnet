{
  _config+:: {
    postgresExporterSelector: 'job=~"$job", instance=~"$instance"',
  },
}
