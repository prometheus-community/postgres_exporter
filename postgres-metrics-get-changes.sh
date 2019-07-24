#!/bin/bash
# Script to parse a text exposition format file into a unique list of metrics
# output by the exporter and then build lists of added/removed metrics.

old_src="$1"
if [ ! -d "$old_src" ] ; then
    mkdir -p "$old_src"
fi

function generate_add_removed() {
    type="$1"
    pg_version="$2"
    old_version="$3"
    new_version="$4"
    
    if [ ! -e "$old_version" ] ; then
        touch "$old_version"
    fi

    comm -23 "$old_version" "$new_version" > ".metrics.${type}.${pg_version}.removed"
    comm -13 "$old_version" "$new_version" > ".metrics.${type}.${pg_version}.added"   
}

for raw_prom in $(echo .*.prom) ; do
    # Get the type and version
    type=$(echo "$raw_prom" | cut -d'.' -f3)
    pg_version=$(echo "$raw_prom" | cut -d'.' -f4- | sed 's/\.prom$//g')

    unique_file="${raw_prom}.unique"
    old_unique_file="$old_src/$unique_file"

    # Strip, sort and deduplicate the label names
    grep -v '#' "$raw_prom" | \
        rev | cut -d' ' -f2- | \
        rev | cut -d'{' -f1 | \
        sort | \
        uniq > "$unique_file"
    
    generate_add_removed "$type" "$pg_version" "$old_unique_file" "$unique_file"
done
