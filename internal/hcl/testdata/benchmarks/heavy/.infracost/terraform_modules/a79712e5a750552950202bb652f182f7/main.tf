locals {
  // If no maps are passed in, we have to provide a default empty map for this module to work with
  input_maps = concat(var.maps, length(var.maps) > 0 ? [] : [{}])
  // Find the greatest depth through the maps
  greatest_depth = max(concat([1], [
    for mod in local.modules :
    concat([
      for i in range(0, length(local.input_maps)) :
      [
        for f in mod[i].fields :
        length(f["path"])
      ]
    ]...)
  ]...)...)

  // For each input map, convert it to a single-level map with a unique key for every nested value
  fields_json = [
    for i in range(0, length(local.input_maps)) :
    merge([
      for j in range(0, local.greatest_depth) :
      {
        for f in local.modules[j][i].fields :
        jsonencode(f.path) => f
      }
    ]...)
  ]

  // Merge the maps using the standard merge function, which will cause higher-precedence map values to overwrite lower-precedence values
  merged_map = merge(local.fields_json...)

  // Split the merged fields into segments for each depth
  merged_fields_by_depth = {
    for depth in range(0, local.greatest_depth) :
    depth => {
      for key in keys(local.merged_map) :
      key => local.merged_map[key]
      if length(local.merged_map[key].path) == depth + 1
    }
  }
  // The lowest level of the re-assembled map is special and not part of the auto-generated depth.tf file
  m0 = {
    for field in local.merged_fields_by_depth[0] :
    field.path[0] => { final_val = field.value, sub_val = lookup(local.m1, field.key, null) }[field.is_final ? "final_val" : "sub_val"]
  }
}

// Check to make sure the highest level module has no remaining values that weren't recursed through
module "asset_sufficient_levels" {
  source        = "Invicton-Labs/assertion/null"
  version       = "0.2.1"
  error_message = "Deepmerge has recursed to insufficient depth (${length(local.modules)} levels is not enough)"
  condition = concat([
    for i in range(0, length(local.input_maps)) :
    local.modules[length(local.modules) - 1][i].remaining
  ]...) == []
}

// Use this  from a DIFFERENT terraform project to generate a new file with a different max depth
/*
resource "local_file" "depth" {
    content     = templatefile("${path.module}/../deepmerge/depth.tmpl", {max_depth = 100})
    filename = "${path.module}/../deepmerge/depth.tf"
}
*/
