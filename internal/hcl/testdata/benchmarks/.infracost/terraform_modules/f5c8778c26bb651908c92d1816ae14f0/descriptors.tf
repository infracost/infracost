# It would be nice to have a fixed array of arguments passed into
# `format()` so all you need to provide is a format string, but
# unfortunately, that does not work easily
# due to https://github.com/hashicorp/terraform/issues/28558
# which requires that the format string consume the last argument passed in.
# We could hack around it by adding then removing a trailing arg, like
#
#   trimsuffix(format("${var.format_string}%[${length(local.labels)+1}]v", concat(local.labels, ["x"])...), "x")
#
# but that is kind of a hack, and overlooks the fact that local.labels
# drops empty label elements, so the index of an element is not guaranteed.
#
#
# So we require the user to specify the arguments as well as the format string.
#

# There is a lot of room for enhancement, but since this is a new feature
# with only 2 use cases, we are going to keep it simple for now.

locals {
  descriptor_labels = { for k, v in local.descriptor_formats : k => [
    for label in v.labels : local.id_context[label]
  ] }
  descriptors = { for k, v in local.descriptor_formats : k => (
    format(v.format, local.descriptor_labels[k]...)
    )
  }
}
