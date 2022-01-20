package version

versions: {
  if ReleaseVersion == "latest" {
    edge: "latest"
    control: "latest"
    catalog: "latest"
    dashboard: "latest"
    jwtsecurity: "latest"
  }
  if ReleaseVersion == "1.7" {
    edge: *"image:tag" | string
    control: "1.7.0"
    catalog: "3.0.0"
    dashboard: "6.0.0"
    jwtsecurity: "1.3.0"
  }
  if ReleaseVersion == "1.6" {
    edge: "1.6.3"
    control: "1.6.5"
    catalog: "2.0.1"
    dashboard: "5.1.1"
    jwtsecurity: "1.3.0"
  }
}