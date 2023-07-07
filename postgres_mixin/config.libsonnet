{
  _config+:: {
    postgresExporterSelector: if self.enableMultiCluster then self.jobSelector + ', ' + self.instance + '=~\"$instance\",cluster=~\"$cluster\"' else  self.jobSelector + ', ' + self.instance + '=~\"$instance\"',
    enableMultiCluster: true,
    instanceLabel: 'instance', //replace to 'pod', in case you are monitoring it on k8s and using pod name as instance selector
    jobSelector: 'job=~\"$job\"',
  },
}
