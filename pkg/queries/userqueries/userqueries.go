package userqueries

func parseUserQueries(content []byte) (map[string]intermediateMetricMap, map[string]string, error) {
	var userQueries UserQueries

	err := yaml.Unmarshal(content, &userQueries)
	if err != nil {
		return nil, nil, err
	}

	// Stores the loaded map representation
	metricMaps := make(map[string]intermediateMetricMap)
	newQueryOverrides := make(map[string]string)

	for metric, specs := range userQueries {
		log.Debugln("New user metric namespace from YAML:", metric, "Will cache results for:", specs.CacheSeconds)
		newQueryOverrides[metric] = specs.Query
		metricMap, ok := metricMaps[metric]
		if !ok {
			// Namespace for metric not found - add it.
			newMetricMap := make(map[string]ColumnMapping)
			metricMap = intermediateMetricMap{
				columnMappings: newMetricMap,
				master:         specs.Master,
				cacheSeconds:   specs.CacheSeconds,
			}
			metricMaps[metric] = metricMap
		}
		for _, metric := range specs.Metrics {
			for name, mappingOption := range metric {
				var columnMapping ColumnMapping
				tmpUsage, _ := stringToColumnUsage(mappingOption.Usage)
				columnMapping.usage = tmpUsage
				columnMapping.description = mappingOption.Description

				// TODO: we should support cu
				columnMapping.mapping = nil
				// Should we support this for users?
				columnMapping.supportedVersions = nil

				metricMap.columnMappings[name] = columnMapping
			}
		}
	}
	return metricMaps, newQueryOverrides, nil
}
