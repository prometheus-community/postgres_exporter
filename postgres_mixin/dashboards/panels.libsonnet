local g = import 'g.libsonnet';

{
  stat: {
    local stat = g.panel.stat,

    base(title, targets):
      stat.new(title)
      +stat.queryOptions.withTargets(targets),

    qps: self.base,
  },

  timeSeries: {
    local timeSeries = g.panel.timeSeries,

    base(title, targets):
      timeSeries.new(title)
      +timeSeries.queryOptions.withTargets(targets),

    ratio(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('percentunit'),

    ratio1(title, targets):
      self.ratio(title, targets)
      + timeSeries.standardOptions.withUnit('percentunit')
      + timeSeries.standardOptions.withMax(1)
  }
}
