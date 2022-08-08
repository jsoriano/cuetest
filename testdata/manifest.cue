spec_version: string

#Version: string
#VersionCondition: string
#Subscription: "basic" | "gold" | "platinum" | "enterprise"
#License: "Apache-2.0" | "Elastic-2.0"

BaseManifest: {
  format_version: #Version
  name: string
  version: #Version
  conditions?: {
    kibana?: version?: #VersionCondition
  }

  // Old format for required subscription, removing it in 2.0.
  if spec_version < "2.0.0" {
    license: #Subscription
  }

  // New format for source license, optional.
  source?: license?: #License

  // New format for required subscription, supported since 1.14.1.
  if spec_version >= "1.14.1" {
    conditions?: elastic?: subscription?: #Subscription
  }

  // Disallow additional contents starting on 2.0.0.
  if spec_version < "2.0.0" {
    ...
  }
}

IntegrationManifest: BaseManifest
IntegrationManifest: type: "integration"

InputManifest: BaseManifest
InputManifest: type: "input"
InputManifest: format_version: >="1.13.0"

InputPackage: {
  "manifest.yml": InputManifest
}

IntegrationPackage: {
  "manifest.yml": IntegrationManifest
}

#PackageManifest: IntegrationManifest | InputManifest
#Package: IntegrationPackage | InputPackage
