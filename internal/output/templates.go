package output

var HTMLTemplate = `
{{define "style"}}
body {
  margin: 0;
  padding: 0.5rem 1rem;
  font-family: sans-serif;
  color: #111827;
}

a {
  color: #3b82f6;
}

.metadata {
  margin-bottom: 1.5rem;
}

.metadata ul {
  list-style-type: none;
  padding: 0;
}

.metadata ul li {
  margin-bottom: 0.5rem;
}

.metadata .label {
  display: inline-block;
  font-weight: bold;
  margin-right: 0.5rem;
  width: 8rem;
}

.warnings {
  margin-top: 1.5rem;
}

table {
  border-collapse: collapse;
  border: 1px solid #6b7280;
}

th, td {
  padding: 0.25rem 0.5rem;
  text-align: left;
}

td.name {
  max-width: 32rem;
}

td.monthly-quantity, td.price, td.hourly-cost, td.monthly-cost {
  text-align: right;
}

tr.group {
  background-color: #e0e7ff;
}

tr.resource {
  background-color: #e5e7eb;
}

tr.resource.top-level {
  background-color: #6b7280;
  color: #ffffff;
}

tr.tags {
  background-color: #6b7280;
  color: #ffffff;
  font-size: 0.75rem;
}

tr.tags td {
  padding-top: 0;
}

tr.total {
  background-color: #ffdfb9;
  font-weight: bold;
}

tr.total td {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.arrow {
  color: #96a0b5;
}
{{end}}

{{define "faviconBase64"}}
iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAMAAACdt4HsAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAABIUExURUdwTBwgLBsgKxwgKyAgLBsgLBwgLBwgLBwgLBwgLBwgLBggKBsgLBwgLRwgKxsgLBAgICAgMBsgKh0gLR0gKhkgLRwgLBwgLEksIvsAAAAXdFJOUwBg36AgkO9wQMCAINDfsOAQEJBQMFB/w99K+AAAAXlJREFUWMPtl8mShCAQRNkLZBOXqf//0zkY0xMIKIQxlwnz2HY+u60lkZBXr9oSFHhw3iMiovLOhAX03Oed98UprIoF2O7cYPBajOu23S4KO8Ro4+4Se1VFWIYD+nroR+RnwKAfEXK/HvWjywHTMAD/HQCeAh5XgbhRwLmbhR/zy3KH+Gd+QlL/NPrGThC8y27o1S6820hKVu6efZRW2YKoAMdXrc0BRUnItk7SmBgPX4yGS1jT7+6YToDKgrmQZVgAkNtuPyisAHoRMzDEAuB+ynPH0D97X7dG0S17I8DmDcInNfz5atZAyiyw600cJCG2nS4863MvCv7QPqn4CaH9oyRT/U/2IZQUF8+Y3zEMpGZ3HRXcIDQgysk1NQeIkIl9eiBpkMG4eBQtRhc40OKHQ7HWmR2ZB6jkgoL+cTL1XGC0y655OxcuT0D5KaqZC5+1Uz3Amc5ccGECrcXx7GchNAUe1JsLF7lARnPhXpVcoGt632Je/ZW+AXMOiOq93TlOAAAAAElFTkSuQmCC
{{end}}

{{define "groupRow"}}
  <tr class="group">
  <td class="name">
    {{.GroupLabel}}: {{.Group}}
  </td>
  <td class="monthly-quantity"></td>
  <td class="unit"></td>
  <td class="price"></td>
  <td class="hourly-cost"></td>
  <td class="monthly-cost"></td>
  </tr>
{{end}}

{{define "resourceRows"}}
  <tr class="resource{{if eq .Indent 0}} top-level{{end}}">
    <td class="name">
      {{if gt .Indent 1}}{{repeat (int (add .Indent -1)) "&nbsp;&nbsp;&nbsp;&nbsp;" | safeHTML}}{{end}}
      {{if gt .Indent 0}}<span class="arrow">&#8627;</span>{{end}}
      {{.Resource.Name}}
    </td>
    <td class="monthly-quantity"></td>
    <td class="unit"></td>
    <td class="price"></td>
    <td class="hourly-cost">{{.Resource.HourlyCost | formatCost}}</td>
    <td class="monthly-cost">{{.Resource.MonthlyCost | formatCost}}</td>
  </tr>
  {{ if .Resource.Tags}}
    <tr class="tags">
      <td class="name">
        {{$tags := list}}
        {{range $k, $v := .Resource.Tags}}
          {{$t := list $k "=" $v | join "" }}
          {{$tags = append $tags $t}}
        {{end}}
        <span class="label">Tags:</span>
        <span>{{$tags | join ", "}}</span>
      </td>
      <td class="monthly-quantity"></td>
      <td class="unit"></td>
      <td class="price"></td>
      <td class="hourly-cost"></td>
      <td class="monthly-cost"></td>
    </tr>
  {{end}}
  {{$ident := add .Indent 1}}
  {{range .Resource.CostComponents}}
    {{template "costComponentRow" dict "CostComponent" . "Indent" $ident}}
  {{end}}
  {{range .Resource.SubResources}}
    {{template "resourceRows" dict "Resource" . "Indent" $ident}}
  {{end}}
{{end}}

{{define "costComponentRow"}}
  <tr class="cost-component">
    <td class="name">
      {{if gt .Indent 1}}{{repeat (int (add .Indent -1)) "&nbsp;&nbsp;&nbsp;&nbsp;" | safeHTML}}{{end}}
      {{if gt .Indent 0}}<span class="arrow">&#8627;</span>{{end}}
      {{.CostComponent.Name}}
    </td>
    <td class="monthly-quantity">{{.CostComponent.MonthlyQuantity | formatQuantity }}</td>
    <td class="unit">{{.CostComponent.Unit}}</td>
    <td class="price">{{.CostComponent.Price | formatAmount }}</td>
    <td class="hourly-cost">{{.CostComponent.HourlyCost | formatCost}}</td>
    <td class="monthly-cost">{{.CostComponent.MonthlyCost | formatCost}}</td>
  </tr>
{{end}}

<!doctype html>
<html>
  <head>
    <title>Infracost cost report</title>
    <style>
      {{template "style"}}
    </style>
    <link id="favicon" rel="shortcut icon" type="image/png" href="data:image/png;base64,{{template "faviconBase64"}}">
  </head>

  <body>
    <div class="metadata">
      <ul>
        <li>
          <span class="label">Generated by:</span>
          <span class="value"><a href="https://infracost.io" target="_blank">Infracost</a></span>
        </li>
        <li>
          <span class="label">Time generated:</span>
          <span class="value">{{.Root.TimeGenerated | date "2006-01-02 15:04:05 MST"}}</span>
        </li>
      </ul>
    </div>

    <table>
      <thead>
        <th class="name">Name</th>
        <th class="monthly-quantity">Monthly quantity</th>
        <th class="unit">Unit</th>
        <th class="price">Price</th>
        <th class="hourly-cost">Hourly cost</th>
        <th class="monthly-cost">Monthly cost</th>
      </thead>
      <tbody>
        {{$groupLabel := .Options.GroupLabel}}
        {{$groupKey := .Options.GroupKey}}
        {{$prevGroup := ""}}
        {{range .Root.Resources}}
          {{$group := index .Metadata $groupKey}}
          {{if ne $group $prevGroup}}
            {{template "groupRow" dict "GroupLabel" $groupLabel "Group" $group}}
          {{end}}
          {{template "resourceRows" dict "Resource" . "Indent" 0}}
          {{$prevGroup = $group}}
        {{end}}
        <tr class="spacer"><td colspan="6"></td></tr>
        <tr class="total">
          <td class="name">Overall total (USD)</td>
          <td class="monthly-quantity"></td>
          <td class="unit"></td>
          <td class="price"></td>
          <td class="hourly-cost">{{.Root.TotalHourlyCost | formatCost}}</td>
          <td class="monthly-cost">{{.Root.TotalMonthlyCost | formatCost}}</td>
        </tr>
      </tbody>
    </table>

    <div class="warnings">
      <p>{{.UnsupportedResourcesMessage | replaceNewLines}}</p>
    </div>
  </body>
</html>`
