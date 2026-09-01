package rest

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterSwaggerRoute registers:
// - /swags: User / Public Developer Swagger Documentation
// - /swagss: Admin Management Swagger Documentation
func RegisterSwaggerRoute(app fiber.Router) {
	// 1. User Swagger UI
	app.Get("/swags", func(c fiber.Ctx) error {
		c.Type("html")
		return c.SendString(userSwaggerHTML)
	})

	app.Get("/swags/openapi.json", func(c fiber.Ctx) error {
		c.Type("json")
		return c.SendString(userOpenAPIJSON)
	})

	// 2. Admin Swagger UI
	app.Get("/swagss", func(c fiber.Ctx) error {
		c.Type("html")
		return c.SendString(adminSwaggerHTML)
	})

	app.Get("/swagss/openapi.json", func(c fiber.Ctx) error {
		c.Type("json")
		return c.SendString(adminOpenAPIJSON)
	})
}

// _____________________________________________________________________________________________________________________
// User / Developer Swagger UI HTML & OpenAPI Specification

const userSwaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>API — Developer Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    html { box-sizing: border-box; }
    *, *:before, *:after { box-sizing: inherit; }
    body {
      margin: 0;
      background: #f8fafc;
      color: #0f172a;
      font-family: 'Inter', sans-serif;
    }
    .topbar { display: none !important; }
    .swagger-ui {
      max-width: 1200px;
      margin: 0 auto;
      padding: 24px 20px;
    }
    .swagger-ui .info .title {
      font-family: 'Inter', sans-serif;
      color: #0f172a;
      font-weight: 700;
    }
    .swagger-ui .info p {
      font-size: 14px;
      color: #475569;
    }
    .swagger-ui .opblock {
      border-radius: 10px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      border: 1px solid #e2e8f0;
      margin-bottom: 12px;
    }
    .swagger-ui .opblock .opblock-summary {
      padding: 10px 16px;
    }
    .swagger-ui .btn.authorize {
      border-color: #06b6d4;
      color: #0891b2;
      border-radius: 8px;
    }
    .swagger-ui .btn.authorize svg {
      fill: #0891b2;
    }
    .swagger-ui .btn.execute {
      background-color: #06b6d4;
      border-color: #06b6d4;
      color: #ffffff;
      border-radius: 8px;
    }
    .swagger-ui code {
      font-family: 'JetBrains Mono', monospace;
    }
    .custom-header {
      background: #ffffff;
      border-bottom: 1px solid #e2e8f0;
      padding: 14px 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: 0 1px 2px rgba(0,0,0,0.03);
    }
    .custom-header img {
      height: 32px;
      object-fit: contain;
    }
    .custom-header a {
      color: #0891b2;
      text-decoration: none;
      font-size: 13px;
      font-weight: 600;
      padding: 6px 12px;
      border-radius: 6px;
      border: 1px solid #e2e8f0;
      background: #f8fafc;
      transition: all 0.15s ease;
    }
    .custom-header a:hover {
      background: #ecfeff;
      border-color: #06b6d4;
    }
  </style>
</head>
<body>
  <div class="custom-header">
    <div style="display:flex;align-items:center;gap:12px;">
      <img src="https://id-cdn.aisbirnusantara.com/documentation/logoneww.png" alt="Aisbir">
      <span style="font-weight:600;font-size:14px;color:#0f172a;">Developer API Documentation</span>
    </div>
    <a href="/">← Return to Dashboard</a>
  </div>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/swags/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

const userOpenAPIJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Developer API",
    "description": "Multi-Device WhatsApp SaaS API for automated messaging, instance management, and webhooks.\n\n### Authentication\nAuthenticate your requests using:\n1. **` + "`" + `X-API-Key` + "`" + `** header or query param **` + "`" + `?apikey=YOUR_KEY` + "`" + `**\n2. **` + "`" + `X-Device-Id` + "`" + `** header or query param **` + "`" + `?device_id=YOUR_DEVICE` + "`" + `**",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "/",
      "description": "API Server"
    }
  ],
  "components": {
    "securitySchemes": {
      "ApiKeyAuth": {
        "type": "apiKey",
        "in": "header",
        "name": "X-API-Key",
        "description": "Developer API Key (X-API-Key header or ?apikey= query)"
      },
      "DeviceIdHeader": {
        "type": "apiKey",
        "in": "header",
        "name": "X-Device-Id",
        "description": "Target WhatsApp Device Identifier (X-Device-Id header or ?device_id= query)"
      }
    }
  },
  "security": [
    {
      "ApiKeyAuth": []
    }
  ],
  "paths": {
    "/send/message": {
      "post": {
        "summary": "Send WhatsApp Text Message",
        "description": "Sends a plain text WhatsApp message to a phone number.",
        "parameters": [
          { "name": "X-Device-Id", "in": "header", "required": false, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["phone", "message"],
                "properties": {
                  "phone": { "type": "string", "example": "628123456789" },
                  "message": { "type": "string", "example": "Hello from API!" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Message dispatched successfully" },
          "400": { "description": "Broadcast quota exceeded or invalid parameters" }
        }
      }
    },
    "/send/image": {
      "post": {
        "summary": "Send WhatsApp Image",
        "description": "Sends an image with optional caption to a recipient.",
        "parameters": [
          { "name": "X-Device-Id", "in": "header", "required": false, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["phone", "image"],
                "properties": {
                  "phone": { "type": "string", "example": "628123456789" },
                  "caption": { "type": "string", "example": "Check out this image!" },
                  "image": { "type": "string", "format": "binary" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Image dispatched successfully" }
        }
      }
    },
    "/send/file": {
      "post": {
        "summary": "Send Document / PDF File",
        "parameters": [
          { "name": "X-Device-Id", "in": "header", "required": false, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["phone", "file"],
                "properties": {
                  "phone": { "type": "string", "example": "628123456789" },
                  "caption": { "type": "string", "example": "Invoice Document" },
                  "file": { "type": "string", "format": "binary" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Document dispatched successfully" }
        }
      }
    },
    "/devices": {
      "get": {
        "summary": "List WhatsApp Devices",
        "description": "Get all WhatsApp device slots for the authenticated tenant.",
        "responses": {
          "200": { "description": "List of device instances" }
        }
      },
      "post": {
        "summary": "Create Device Slot",
        "description": "Provision a new WhatsApp instance slot within plan device limits.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["device_id"],
                "properties": {
                  "device_id": { "type": "string", "example": "cs-bot-1" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Device slot created" }
        }
      }
    },
    "/devices/{id}/login": {
      "get": {
        "summary": "Get QR Code for Pairing",
        "description": "Retrieves the Base64 QR code image to link a WhatsApp session.",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Base64 QR image data" }
        }
      }
    },
    "/devices/{id}/login/code": {
      "post": {
        "summary": "Pair Device via 8-Digit Code",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["phone"],
                "properties": {
                  "phone": { "type": "string", "example": "628123456789" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "8-digit WhatsApp pairing code" }
        }
      }
    },
    "/devices/{id}/reconnect": {
      "post": {
        "summary": "Reconnect WhatsApp Device Session",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Reconnect initiated" }
        }
      }
    },
    "/devices/{id}/logout": {
      "post": {
        "summary": "Disconnect & Logout WhatsApp Session",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Logged out" }
        }
      }
    },
    "/devices/{id}": {
      "delete": {
        "summary": "Delete Device Instance Slot",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Device slot removed" }
        }
      }
    },
    "/chats": {
      "get": {
        "summary": "List Synced WhatsApp Chats",
        "parameters": [
          { "name": "X-Device-Id", "in": "header", "required": false, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "List of recent conversations" }
        }
      }
    },
    "/api/user/api-keys": {
      "get": {
        "summary": "List API Keys",
        "responses": {
          "200": { "description": "List of developer API keys" }
        }
      },
      "post": {
        "summary": "Generate API Key",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name"],
                "properties": {
                  "name": { "type": "string", "example": "Backend Production" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "API key generated" }
        }
      }
    },
    "/api/user/api-keys/{id}": {
      "delete": {
        "summary": "Delete API Key",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "responses": {
          "200": { "description": "API key deleted" }
        }
      }
    },
    "/api/user/webhooks": {
      "get": {
        "summary": "List Configured Outbound Webhooks",
        "responses": {
          "200": { "description": "List of user webhooks" }
        }
      },
      "post": {
        "summary": "Create Outbound Webhook",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["url"],
                "properties": {
                  "url": { "type": "string", "example": "https://yourdomain.com/webhook" },
                  "secret": { "type": "string", "example": "secret_token_123" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Webhook configured" }
        }
      }
    },
    "/api/plans": {
      "get": {
        "summary": "List Available Plans & Quotas",
        "responses": {
          "200": { "description": "List of active plans" }
        }
      }
    },
    "/api/user/select-plan": {
      "post": {
        "summary": "Select Subscription Plan",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["plan_id"],
                "properties": {
                  "plan_id": { "type": "string", "example": "ocean" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Plan selected" }
        }
      }
    },
    "/api/user/auto-responses": {
      "get": {
        "summary": "List Auto Response Rules",
        "description": "Get all configured chatbot auto-response trigger rules.",
        "responses": {
          "200": { "description": "List of auto-response rules" }
        }
      },
      "post": {
        "summary": "Create Auto Response Rule",
        "description": "Configure an automated keyword or command trigger with dynamic template variables ({name}, {sender}, {time}, {date}).",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["trigger_keyword", "response_message"],
                "properties": {
                  "device_id": { "type": "string", "example": "okebolo" },
                  "trigger_type": { "type": "string", "enum": ["exact", "contains", "starts_with", "regex"], "default": "contains" },
                  "trigger_keyword": { "type": "string", "example": "!menu" },
                  "response_message": { "type": "string", "example": "Halo {name}! Welcome to Aisbir. Available commands: !price, !info, !help" },
                  "is_active": { "type": "boolean", "default": true }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Auto response created" }
        }
      }
    },
    "/api/user/auto-responses/{id}": {
      "put": {
        "summary": "Update Auto Response Rule",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "device_id": { "type": "string" },
                  "trigger_type": { "type": "string", "enum": ["exact", "contains", "starts_with", "regex"] },
                  "trigger_keyword": { "type": "string" },
                  "response_message": { "type": "string" },
                  "is_active": { "type": "boolean" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Rule updated" }
        }
      },
      "delete": {
        "summary": "Delete Auto Response Rule",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "responses": {
          "200": { "description": "Rule deleted" }
        }
      }
    }
  }
}`

// _____________________________________________________________________________________________________________________
// Admin Management Swagger UI HTML & OpenAPI Specification

const adminSwaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>SaaS — Admin Management API</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    html { box-sizing: border-box; }
    *, *:before, *:after { box-sizing: inherit; }
    body {
      margin: 0;
      background: #f8fafc;
      color: #0f172a;
      font-family: 'Inter', sans-serif;
    }
    .topbar { display: none !important; }
    .swagger-ui {
      max-width: 1200px;
      margin: 0 auto;
      padding: 24px 20px;
    }
    .swagger-ui .info .title {
      font-family: 'Inter', sans-serif;
      color: #0f172a;
      font-weight: 700;
    }
    .swagger-ui .info p {
      font-size: 14px;
      color: #475569;
    }
    .swagger-ui .opblock {
      border-radius: 10px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      border: 1px solid #e2e8f0;
      margin-bottom: 12px;
    }
    .swagger-ui .opblock .opblock-summary {
      padding: 10px 16px;
    }
    .swagger-ui .btn.authorize {
      border-color: #06b6d4;
      color: #0891b2;
      border-radius: 8px;
    }
    .swagger-ui .btn.authorize svg {
      fill: #0891b2;
    }
    .swagger-ui .btn.execute {
      background-color: #06b6d4;
      border-color: #06b6d4;
      color: #ffffff;
      border-radius: 8px;
    }
    .swagger-ui code {
      font-family: 'JetBrains Mono', monospace;
    }
    .custom-header {
      background: #ffffff;
      border-bottom: 1px solid #e2e8f0;
      padding: 14px 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: 0 1px 2px rgba(0,0,0,0.03);
    }
    .custom-header img {
      height: 32px;
      object-fit: contain;
    }
    .admin-badge {
      background: #ecfeff;
      color: #0891b2;
      font-weight: 700;
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 4px;
      border: 1px solid #a5f3fc;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .custom-header a {
      color: #0891b2;
      text-decoration: none;
      font-size: 13px;
      font-weight: 600;
      padding: 6px 12px;
      border-radius: 6px;
      border: 1px solid #e2e8f0;
      background: #f8fafc;
      transition: all 0.15s ease;
    }
    .custom-header a:hover {
      background: #ecfeff;
      border-color: #06b6d4;
    }
  </style>
</head>
<body>
  <div class="custom-header">
    <div style="display:flex;align-items:center;gap:12px;">
      <img src="https://id-cdn.aisbirnusantara.com/documentation/logoneww.png" alt="Aisbir">
      <span style="font-weight:600;font-size:14px;color:#0f172a;">Admin Management API</span>
      <span class="admin-badge">Admin Only</span>
    </div>
    <div style="display:flex;gap:10px;">
      <a href="/swags">User Docs ↗</a>
      <a href="/">← Dashboard</a>
    </div>
  </div>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/swagss/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

const adminOpenAPIJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "SaaS — Admin Management API",
    "description": "Administrative endpoints for managing SaaS tenants, users, subscriptions, quotas, and pricing plans.\n\n### Authentication\nRequires Admin JWT token (**` + "`" + `Bearer <token>` + "`" + `**) or Admin API Key (**` + "`" + `X-API-Key` + "`" + `**).",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "/",
      "description": "API Server"
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Admin JWT Token (Bearer <token>)"
      },
      "ApiKeyAuth": {
        "type": "apiKey",
        "in": "header",
        "name": "X-API-Key",
        "description": "Admin API Key (X-API-Key header)"
      }
    }
  },
  "security": [
    {
      "BearerAuth": []
    },
    {
      "ApiKeyAuth": []
    }
  ],
  "paths": {
    "/api/admin/users": {
      "get": {
        "summary": "List All SaaS Tenants",
        "description": "Returns all registered user accounts with their active tier, expiry dates, and usage stats.",
        "responses": {
          "200": { "description": "List of user accounts" }
        }
      },
      "post": {
        "summary": "Create / Provision New User",
        "description": "Provision a new tenant user account with assigned role and subscription tier.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["full_name", "email", "password"],
                "properties": {
                  "full_name": { "type": "string", "example": "Customer Company" },
                  "email": { "type": "string", "example": "customer@example.com" },
                  "password": { "type": "string", "example": "securepassword123" },
                  "role": { "type": "string", "enum": ["pelanggan", "admin"], "default": "pelanggan" },
                  "tier_id": { "type": "string", "example": "ocean" },
                  "duration_days": { "type": "integer", "example": 30 }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "User created successfully" }
        }
      }
    },
    "/api/admin/users/{id}/plan": {
      "put": {
        "summary": "Change User Tier Plan",
        "description": "Assigns a different subscription plan (e.g. Ocean, Sea, or custom tier) to a user.",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["tier_id"],
                "properties": {
                  "tier_id": { "type": "string", "example": "sea" },
                  "duration_days": { "type": "integer", "example": 30 }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "User plan updated" }
        }
      }
    },
    "/api/admin/users/{id}/renew": {
      "post": {
        "summary": "Renew / Extend User Tier Validity",
        "description": "Extends the validity period of a user's subscription tier and resets their broadcast counter.",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["duration_days"],
                "properties": {
                  "duration_days": { "type": "integer", "example": 30 }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Tier validity extended" }
        }
      }
    },
    "/api/admin/users/{id}": {
      "delete": {
        "summary": "Delete User Account",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }
        ],
        "responses": {
          "200": { "description": "User deleted" }
        }
      }
    },
    "/api/admin/plans": {
      "get": {
        "summary": "List All Configured Plans",
        "description": "Returns all subscription plans with device limits, broadcast caps, API keys, and pricing.",
        "responses": {
          "200": { "description": "List of plans" }
        }
      },
      "post": {
        "summary": "Create New Subscription Plan",
        "description": "Define a new subscription plan with customized quotas and pricing.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["id", "name", "device_limit", "broadcast_limit"],
                "properties": {
                  "id": { "type": "string", "example": "enterprise" },
                  "name": { "type": "string", "example": "Enterprise Scale" },
                  "device_limit": { "type": "integer", "example": 10 },
                  "broadcast_limit": { "type": "integer", "example": -1 },
                  "api_key_limit": { "type": "integer", "example": 5 },
                  "webhook_limit": { "type": "integer", "example": 5 },
                  "duration_days": { "type": "integer", "example": 30 },
                  "price_monthly": { "type": "integer", "example": 499000 },
                  "description": { "type": "string", "example": "Enterprise tier for high volume workloads." }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Plan created successfully" }
        }
      }
    },
    "/api/admin/plans/{id}": {
      "put": {
        "summary": "Update Plan Configuration",
        "description": "Modifies device limits, broadcast limits, price, or description for an existing plan.",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": { "type": "string" },
                  "device_limit": { "type": "integer" },
                  "broadcast_limit": { "type": "integer" },
                  "api_key_limit": { "type": "integer" },
                  "webhook_limit": { "type": "integer" },
                  "duration_days": { "type": "integer" },
                  "price_monthly": { "type": "integer" },
                  "description": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Plan updated successfully" }
        }
      },
      "delete": {
        "summary": "Delete Subscription Plan",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Plan deleted" }
        }
      }
    }
  }
}`
