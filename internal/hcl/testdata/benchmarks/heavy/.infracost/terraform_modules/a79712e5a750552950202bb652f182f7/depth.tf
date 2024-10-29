locals {
  mod0toparse = [
    for map_idx in range(0, length(local.input_maps)) :
    [{
      path  = [],
      value = local.input_maps[map_idx]
    }]
  ]
  mod0 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod0toparse[map_idx] :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod0toparse[map_idx] :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod1 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod0[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod0[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod2 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod1[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod1[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod3 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod2[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod2[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod4 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod3[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod3[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod5 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod4[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod4[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod6 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod5[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod5[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod7 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod6[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod6[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod8 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod7[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod7[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod9 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod8[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod8[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod10 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod9[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod9[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod11 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod10[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod10[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod12 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod11[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod11[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod13 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod12[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod12[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod14 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod13[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod13[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod15 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod14[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod14[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod16 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod15[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod15[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod17 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod16[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod16[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod18 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod17[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod17[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod19 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod18[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod18[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod20 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod19[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod19[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod21 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod20[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod20[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod22 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod21[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod21[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod23 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod22[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod22[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod24 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod23[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod23[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod25 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod24[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod24[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod26 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod25[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod25[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod27 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod26[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod26[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod28 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod27[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod27[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod29 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod28[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod28[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod30 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod29[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod29[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod31 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod30[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod30[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod32 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod31[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod31[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod33 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod32[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod32[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod34 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod33[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod33[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod35 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod34[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod34[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod36 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod35[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod35[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod37 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod36[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod36[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod38 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod37[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod37[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod39 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod38[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod38[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod40 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod39[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod39[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod41 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod40[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod40[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod42 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod41[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod41[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod43 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod42[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod42[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod44 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod43[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod43[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod45 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod44[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod44[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod46 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod45[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod45[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod47 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod46[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod46[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod48 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod47[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod47[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod49 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod48[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod48[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod50 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod49[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod49[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod51 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod50[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod50[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod52 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod51[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod51[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod53 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod52[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod52[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod54 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod53[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod53[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod55 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod54[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod54[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod56 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod55[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod55[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod57 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod56[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod56[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod58 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod57[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod57[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod59 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod58[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod58[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod60 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod59[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod59[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod61 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod60[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod60[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod62 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod61[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod61[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod63 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod62[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod62[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod64 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod63[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod63[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod65 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod64[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod64[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod66 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod65[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod65[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod67 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod66[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod66[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod68 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod67[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod67[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod69 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod68[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod68[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod70 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod69[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod69[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod71 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod70[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod70[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod72 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod71[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod71[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod73 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod72[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod72[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod74 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod73[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod73[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod75 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod74[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod74[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod76 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod75[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod75[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod77 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod76[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod76[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod78 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod77[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod77[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod79 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod78[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod78[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod80 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod79[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod79[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod81 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod80[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod80[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod82 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod81[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod81[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod83 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod82[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod82[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod84 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod83[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod83[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod85 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod84[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod84[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod86 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod85[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod85[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod87 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod86[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod86[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod88 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod87[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod87[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod89 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod88[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod88[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod90 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod89[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod89[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod91 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod90[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod90[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod92 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod91[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod91[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod93 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod92[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod92[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod94 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod93[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod93[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod95 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod94[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod94[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod96 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod95[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod95[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod97 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod96[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod96[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod98 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod97[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod97[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod99 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod98[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod98[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]
  mod100 = [
    for map_idx in range(0, length(local.input_maps)) :
    {
      fields = concat([], [
        for item in local.mod99[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            key      = jsonencode(concat(item["path"], [key])),
            path     = concat(item["path"], [key]),
            value    = item["value"][key],
            is_final = item["value"][key] == null || try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) != null
          }
        ]
      ]...)
      remaining = concat([], [
        for item in local.mod99[map_idx].remaining :
        [
          for key in keys(item["value"]) :
          {
            path  = concat(item["path"], [key]),
            value = item["value"][key]
          }
          if item["value"][key] != null && try(tolist(item["value"][key]), toset(item["value"][key]), tonumber(item["value"][key]), tobool(item["value"][key]), tostring(item["value"][key]), null) == null
        ]
      ]...)
    }
  ]

  modules = [
    local.mod0,
    local.mod1,
    local.mod2,
    local.mod3,
    local.mod4,
    local.mod5,
    local.mod6,
    local.mod7,
    local.mod8,
    local.mod9,
    local.mod10,
    local.mod11,
    local.mod12,
    local.mod13,
    local.mod14,
    local.mod15,
    local.mod16,
    local.mod17,
    local.mod18,
    local.mod19,
    local.mod20,
    local.mod21,
    local.mod22,
    local.mod23,
    local.mod24,
    local.mod25,
    local.mod26,
    local.mod27,
    local.mod28,
    local.mod29,
    local.mod30,
    local.mod31,
    local.mod32,
    local.mod33,
    local.mod34,
    local.mod35,
    local.mod36,
    local.mod37,
    local.mod38,
    local.mod39,
    local.mod40,
    local.mod41,
    local.mod42,
    local.mod43,
    local.mod44,
    local.mod45,
    local.mod46,
    local.mod47,
    local.mod48,
    local.mod49,
    local.mod50,
    local.mod51,
    local.mod52,
    local.mod53,
    local.mod54,
    local.mod55,
    local.mod56,
    local.mod57,
    local.mod58,
    local.mod59,
    local.mod60,
    local.mod61,
    local.mod62,
    local.mod63,
    local.mod64,
    local.mod65,
    local.mod66,
    local.mod67,
    local.mod68,
    local.mod69,
    local.mod70,
    local.mod71,
    local.mod72,
    local.mod73,
    local.mod74,
    local.mod75,
    local.mod76,
    local.mod77,
    local.mod78,
    local.mod79,
    local.mod80,
    local.mod81,
    local.mod82,
    local.mod83,
    local.mod84,
    local.mod85,
    local.mod86,
    local.mod87,
    local.mod88,
    local.mod89,
    local.mod90,
    local.mod91,
    local.mod92,
    local.mod93,
    local.mod94,
    local.mod95,
    local.mod96,
    local.mod97,
    local.mod98,
    local.mod99,
    local.mod100,
  ]

  m101 = {}
  m100 = {
    for field in lookup(local.merged_fields_by_depth, 99, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 100, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m101, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m99 = {
    for field in lookup(local.merged_fields_by_depth, 98, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 99, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m100, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m98 = {
    for field in lookup(local.merged_fields_by_depth, 97, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 98, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m99, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m97 = {
    for field in lookup(local.merged_fields_by_depth, 96, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 97, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m98, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m96 = {
    for field in lookup(local.merged_fields_by_depth, 95, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 96, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m97, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m95 = {
    for field in lookup(local.merged_fields_by_depth, 94, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 95, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m96, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m94 = {
    for field in lookup(local.merged_fields_by_depth, 93, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 94, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m95, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m93 = {
    for field in lookup(local.merged_fields_by_depth, 92, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 93, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m94, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m92 = {
    for field in lookup(local.merged_fields_by_depth, 91, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 92, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m93, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m91 = {
    for field in lookup(local.merged_fields_by_depth, 90, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 91, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m92, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m90 = {
    for field in lookup(local.merged_fields_by_depth, 89, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 90, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m91, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m89 = {
    for field in lookup(local.merged_fields_by_depth, 88, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 89, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m90, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m88 = {
    for field in lookup(local.merged_fields_by_depth, 87, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 88, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m89, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m87 = {
    for field in lookup(local.merged_fields_by_depth, 86, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 87, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m88, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m86 = {
    for field in lookup(local.merged_fields_by_depth, 85, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 86, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m87, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m85 = {
    for field in lookup(local.merged_fields_by_depth, 84, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 85, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m86, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m84 = {
    for field in lookup(local.merged_fields_by_depth, 83, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 84, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m85, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m83 = {
    for field in lookup(local.merged_fields_by_depth, 82, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 83, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m84, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m82 = {
    for field in lookup(local.merged_fields_by_depth, 81, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 82, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m83, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m81 = {
    for field in lookup(local.merged_fields_by_depth, 80, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 81, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m82, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m80 = {
    for field in lookup(local.merged_fields_by_depth, 79, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 80, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m81, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m79 = {
    for field in lookup(local.merged_fields_by_depth, 78, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 79, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m80, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m78 = {
    for field in lookup(local.merged_fields_by_depth, 77, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 78, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m79, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m77 = {
    for field in lookup(local.merged_fields_by_depth, 76, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 77, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m78, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m76 = {
    for field in lookup(local.merged_fields_by_depth, 75, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 76, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m77, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m75 = {
    for field in lookup(local.merged_fields_by_depth, 74, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 75, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m76, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m74 = {
    for field in lookup(local.merged_fields_by_depth, 73, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 74, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m75, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m73 = {
    for field in lookup(local.merged_fields_by_depth, 72, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 73, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m74, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m72 = {
    for field in lookup(local.merged_fields_by_depth, 71, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 72, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m73, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m71 = {
    for field in lookup(local.merged_fields_by_depth, 70, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 71, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m72, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m70 = {
    for field in lookup(local.merged_fields_by_depth, 69, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 70, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m71, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m69 = {
    for field in lookup(local.merged_fields_by_depth, 68, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 69, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m70, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m68 = {
    for field in lookup(local.merged_fields_by_depth, 67, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 68, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m69, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m67 = {
    for field in lookup(local.merged_fields_by_depth, 66, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 67, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m68, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m66 = {
    for field in lookup(local.merged_fields_by_depth, 65, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 66, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m67, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m65 = {
    for field in lookup(local.merged_fields_by_depth, 64, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 65, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m66, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m64 = {
    for field in lookup(local.merged_fields_by_depth, 63, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 64, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m65, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m63 = {
    for field in lookup(local.merged_fields_by_depth, 62, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 63, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m64, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m62 = {
    for field in lookup(local.merged_fields_by_depth, 61, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 62, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m63, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m61 = {
    for field in lookup(local.merged_fields_by_depth, 60, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 61, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m62, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m60 = {
    for field in lookup(local.merged_fields_by_depth, 59, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 60, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m61, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m59 = {
    for field in lookup(local.merged_fields_by_depth, 58, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 59, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m60, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m58 = {
    for field in lookup(local.merged_fields_by_depth, 57, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 58, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m59, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m57 = {
    for field in lookup(local.merged_fields_by_depth, 56, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 57, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m58, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m56 = {
    for field in lookup(local.merged_fields_by_depth, 55, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 56, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m57, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m55 = {
    for field in lookup(local.merged_fields_by_depth, 54, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 55, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m56, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m54 = {
    for field in lookup(local.merged_fields_by_depth, 53, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 54, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m55, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m53 = {
    for field in lookup(local.merged_fields_by_depth, 52, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 53, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m54, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m52 = {
    for field in lookup(local.merged_fields_by_depth, 51, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 52, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m53, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m51 = {
    for field in lookup(local.merged_fields_by_depth, 50, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 51, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m52, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m50 = {
    for field in lookup(local.merged_fields_by_depth, 49, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 50, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m51, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m49 = {
    for field in lookup(local.merged_fields_by_depth, 48, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 49, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m50, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m48 = {
    for field in lookup(local.merged_fields_by_depth, 47, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 48, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m49, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m47 = {
    for field in lookup(local.merged_fields_by_depth, 46, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 47, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m48, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m46 = {
    for field in lookup(local.merged_fields_by_depth, 45, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 46, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m47, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m45 = {
    for field in lookup(local.merged_fields_by_depth, 44, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 45, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m46, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m44 = {
    for field in lookup(local.merged_fields_by_depth, 43, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 44, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m45, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m43 = {
    for field in lookup(local.merged_fields_by_depth, 42, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 43, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m44, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m42 = {
    for field in lookup(local.merged_fields_by_depth, 41, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 42, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m43, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m41 = {
    for field in lookup(local.merged_fields_by_depth, 40, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 41, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m42, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m40 = {
    for field in lookup(local.merged_fields_by_depth, 39, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 40, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m41, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m39 = {
    for field in lookup(local.merged_fields_by_depth, 38, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 39, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m40, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m38 = {
    for field in lookup(local.merged_fields_by_depth, 37, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 38, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m39, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m37 = {
    for field in lookup(local.merged_fields_by_depth, 36, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 37, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m38, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m36 = {
    for field in lookup(local.merged_fields_by_depth, 35, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 36, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m37, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m35 = {
    for field in lookup(local.merged_fields_by_depth, 34, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 35, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m36, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m34 = {
    for field in lookup(local.merged_fields_by_depth, 33, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 34, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m35, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m33 = {
    for field in lookup(local.merged_fields_by_depth, 32, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 33, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m34, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m32 = {
    for field in lookup(local.merged_fields_by_depth, 31, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 32, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m33, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m31 = {
    for field in lookup(local.merged_fields_by_depth, 30, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 31, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m32, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m30 = {
    for field in lookup(local.merged_fields_by_depth, 29, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 30, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m31, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m29 = {
    for field in lookup(local.merged_fields_by_depth, 28, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 29, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m30, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m28 = {
    for field in lookup(local.merged_fields_by_depth, 27, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 28, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m29, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m27 = {
    for field in lookup(local.merged_fields_by_depth, 26, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 27, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m28, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m26 = {
    for field in lookup(local.merged_fields_by_depth, 25, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 26, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m27, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m25 = {
    for field in lookup(local.merged_fields_by_depth, 24, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 25, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m26, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m24 = {
    for field in lookup(local.merged_fields_by_depth, 23, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 24, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m25, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m23 = {
    for field in lookup(local.merged_fields_by_depth, 22, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 23, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m24, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m22 = {
    for field in lookup(local.merged_fields_by_depth, 21, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 22, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m23, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m21 = {
    for field in lookup(local.merged_fields_by_depth, 20, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 21, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m22, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m20 = {
    for field in lookup(local.merged_fields_by_depth, 19, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 20, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m21, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m19 = {
    for field in lookup(local.merged_fields_by_depth, 18, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 19, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m20, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m18 = {
    for field in lookup(local.merged_fields_by_depth, 17, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 18, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m19, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m17 = {
    for field in lookup(local.merged_fields_by_depth, 16, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 17, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m18, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m16 = {
    for field in lookup(local.merged_fields_by_depth, 15, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 16, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m17, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m15 = {
    for field in lookup(local.merged_fields_by_depth, 14, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 15, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m16, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m14 = {
    for field in lookup(local.merged_fields_by_depth, 13, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 14, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m15, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m13 = {
    for field in lookup(local.merged_fields_by_depth, 12, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 13, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m14, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m12 = {
    for field in lookup(local.merged_fields_by_depth, 11, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 12, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m13, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m11 = {
    for field in lookup(local.merged_fields_by_depth, 10, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 11, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m12, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m10 = {
    for field in lookup(local.merged_fields_by_depth, 9, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 10, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m11, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m9 = {
    for field in lookup(local.merged_fields_by_depth, 8, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 9, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m10, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m8 = {
    for field in lookup(local.merged_fields_by_depth, 7, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 8, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m9, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m7 = {
    for field in lookup(local.merged_fields_by_depth, 6, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 7, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m8, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m6 = {
    for field in lookup(local.merged_fields_by_depth, 5, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 6, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m7, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m5 = {
    for field in lookup(local.merged_fields_by_depth, 4, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 5, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m6, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m4 = {
    for field in lookup(local.merged_fields_by_depth, 3, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 4, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m5, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m3 = {
    for field in lookup(local.merged_fields_by_depth, 2, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 3, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m4, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m2 = {
    for field in lookup(local.merged_fields_by_depth, 1, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 2, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m3, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
  m1 = {
    for field in lookup(local.merged_fields_by_depth, 0, {}) :
    field.key => { final_val = field.value, sub_val = {
      for subfield in lookup(local.merged_fields_by_depth, 1, {}) :
      subfield.path[length(subfield.path) - 1] => lookup(local.m2, subfield.key, subfield.value)
      if slice(subfield.path, 0, length(subfield.path) - 1) == field.path
    } }[field.is_final ? "final_val" : "sub_val"]
  }
}
