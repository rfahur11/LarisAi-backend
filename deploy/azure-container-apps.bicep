@description('Location for Azure resources')
param location string = 'southeastasia'

@description('Environment prefix for naming')
param environmentName string = 'larisai'

@description('Container image tag or URI')
param containerImage string = 'mcr.microsoft.com/azuredocs/aci-helloworld:latest'

@description('MongoDB Atlas or Cosmos DB Connection URI')
@secure()
param mongoUri string

@description('MongoDB Database Name')
param mongoDbName string = 'larisai_pos'

@description('JWT Secret Key for SaaS Multi-Tenant Token Signing')
@secure()
param jwtSecret string = 'super_secret_larisai_key_2026'

// Resource Names
var logAnalyticsName = '${environmentName}-logs'
var appEnvName = '${environmentName}-ca-env'
var appName = '${environmentName}-pos-api'

// 1. Log Analytics Workspace (Per-GB / Free tier ingestion)
resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2023-09-01' = {
  name: logAnalyticsName
  location: location
  properties: {
    sku: {
      name: 'PerGB2018'
    }
    retentionInDays: 30
  }
}

// 2. Azure Container Apps Managed Environment (Consumption Plan)
resource containerAppEnv 'Microsoft.App/managedEnvironments@2024-03-01' = {
  name: appEnvName
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logAnalytics.properties.customerId
        sharedKey: logAnalytics.listKeys().primarySharedKey
      }
    }
  }
}

// 3. Azure Container App (Scale-to-Zero Consumption Serverless)
resource posContainerApp 'Microsoft.App/containerApps@2024-03-01' = {
  name: appName
  location: location
  properties: {
    managedEnvironmentId: containerAppEnv.id
    configuration: {
      ingress: {
        external: true
        targetPort: 8080
        allowInsecure: false
        transport: 'auto'
      }
      secrets: [
        {
          name: 'mongo-uri'
          value: mongoUri
        }
        {
          name: 'jwt-secret'
          value: jwtSecret
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'pos-service'
          image: containerImage
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
          env: [
            {
              name: 'PORT'
              value: '8080'
            }
            {
              name: 'MONGO_DB_NAME'
              value: mongoDbName
            }
            {
              name: 'MONGO_URI'
              secretRef: 'mongo-uri'
            }
            {
              name: 'JWT_SECRET'
              secretRef: 'jwt-secret'
            }
          ]
        }
      ]
      scale: {
        // FINOPS GUARDRAIL: Scale to Zero when no requests, max 1 replica for student cap protection
        minReplicas: 0
        maxReplicas: 1
        rules: [
          {
            name: 'http-scaling'
            http: {
              metadata: {
                concurrentRequests: '50'
              }
            }
          }
        ]
      }
    }
  }
}

output fqdn string = posContainerApp.properties.configuration.ingress.fqdn
