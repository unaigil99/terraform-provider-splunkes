# Documento de Desarrollo: terraform-provider-splunkes

## Resumen Ejecutivo

Se ha desarrollado un **Terraform Provider completo para Splunk Enterprise Security** (`terraform-provider-splunkes`) usando Go y el moderno `terraform-plugin-framework` (protocolo v6). El proveedor permite gestionar la infraestructura de seguridad de Splunk ES como codigo (Security-as-Code), cubriendo 12 recursos y 7 data sources.

**Estado final:** Compilacion limpia (`go build` + `go vet` sin errores). Binario de 23MB listo para uso.

---

## 1. Arquitectura del Proyecto

### 1.1 Stack Tecnologico

| Componente | Version | Justificacion |
|------------|---------|---------------|
| Go | 1.22.0 | Compatible con el toolchain local (go1.22.1) |
| terraform-plugin-framework | v1.13.0 | Framework moderno (protocolo v6), no el legacy SDK v2 |
| terraform-plugin-framework-validators | v0.16.0 | Validadores para campos de esquema |
| terraform-plugin-log | v0.9.0 | Logging estructurado para depuracion |

### 1.2 Estructura de Archivos

```
terraform-provider-splunkes/
├── main.go                              # Punto de entrada (29 lineas)
├── go.mod / go.sum                      # Dependencias Go
├── GNUmakefile                          # Targets: build, test, testacc, fmt, lint, vet
├── .goreleaser.yml                      # Configuracion de release multiplataforma
├── terraform-registry-manifest.json     # Manifiesto de registro (protocolo v6)
├── README.md                            # Documentacion del proveedor
├── docs/index.md                        # Documentacion para Terraform Registry
│
├── internal/
│   ├── sdk/                             # Capa de cliente HTTP (11 archivos, 948 lineas)
│   │   ├── client.go                    # Cliente HTTP central con auth, rate limiting, retries
│   │   ├── saved_searches.go            # CRUD para saved searches
│   │   ├── macros.go                    # CRUD para macros de busqueda
│   │   ├── lookups.go                   # CRUD para lookup definitions y tables
│   │   ├── kvstore.go                   # CRUD para KV store collections
│   │   ├── investigations.go            # CRUD para ES v2 investigations + notes
│   │   ├── findings.go                  # CRUD para ES v2 findings
│   │   ├── risks.go                     # Lectura/modificacion de risk scores
│   │   ├── threat_intel.go              # CRUD para threat intelligence items
│   │   ├── analytic_stories.go          # CRUD para analytic stories
│   │   └── assets_identity.go           # Lectura de assets e identities (solo lectura)
│   │
│   └── provider/                        # Capa del proveedor Terraform (20 archivos, 5726 lineas)
│       ├── provider.go                  # Registro de recursos y data sources, configuracion
│       ├── resource_correlation_search.go    # 847 lineas - recurso mas complejo
│       ├── resource_saved_search.go          # 415 lineas
│       ├── resource_macro.go                 # 330 lineas
│       ├── resource_lookup_definition.go     # 383 lineas
│       ├── resource_lookup_table.go          # 273 lineas
│       ├── resource_kvstore_collection.go    # 313 lineas
│       ├── resource_investigation.go         # 392 lineas
│       ├── resource_investigation_note.go    # 325 lineas
│       ├── resource_finding.go               # 368 lineas
│       ├── resource_risk_modifier.go         # 290 lineas
│       ├── resource_threat_intel.go          # 300 lineas
│       ├── resource_analytic_story.go        # 359 lineas
│       ├── datasource_correlation_search.go  # 179 lineas
│       ├── datasource_investigation.go       # 117 lineas
│       ├── datasource_finding.go             # 124 lineas
│       ├── datasource_risk_score.go          # 110 lineas
│       ├── datasource_identity.go            # 130 lineas
│       ├── datasource_asset.go               # 136 lineas
│       └── datasource_macro.go               # 154 lineas
│
└── examples/                            # Ejemplos de uso (9 archivos .tf)
    ├── provider/provider.tf
    ├── full_detection_pipeline/main.tf   # Pipeline E2E completo
    └── resources/
        ├── splunkes_correlation_search/resource.tf
        ├── splunkes_macro/resource.tf
        ├── splunkes_investigation/resource.tf
        ├── splunkes_finding/resource.tf
        ├── splunkes_threat_intel/resource.tf
        └── splunkes_risk_modifier/resource.tf
```

**Total:** 32 archivos Go, 6.703 lineas de codigo.

### 1.3 Arquitectura de Referencia

La arquitectura se inspiro en el proveedor `fortinetdev/terraform-provider-fortisase`, adaptando los siguientes patrones:

- **SDK interno separado del provider** (`internal/sdk/` vs `internal/provider/`)
- **Mutex per-resource-type** para evitar race conditions en operaciones concurrentes
- **Rate limiting** de 200ms entre peticiones HTTP
- **Estructura plana** en `provider/` (un archivo por recurso/data source)

---

## 2. Capa SDK (`internal/sdk/`)

### 2.1 Cliente HTTP Central (`client.go`)

El cliente implementa las siguientes caracteristicas:

#### Autenticacion (3 metodos, en orden de prioridad)
1. **Bearer Token** (`Authorization: Bearer <token>`)
2. **Session Key** (`Authorization: Splunk <key>`)
3. **Basic Auth** (`username:password`)

#### Resiliencia
- **Rate Limiting:** Canal de Go con tick cada 200ms, garantiza max 5 req/s
- **Reintentos:** Hasta 5 intentos con backoff exponencial para errores 429 (rate limit) y 5xx (server error)
- **Mutex per-recurso:** Previene condiciones de carrera en operaciones concurrentes sobre el mismo tipo de recurso

#### Dos metodos de peticion HTTP

| Metodo | Uso | Content-Type | Body | output_mode |
|--------|-----|-------------|------|-------------|
| `DoRequest` | API REST tradicional de Splunk | `application/x-www-form-urlencoded` | `url.Values` | Se anade automaticamente |
| `DoJSONRequest` | API ES v2 | `application/json` | JSON marshaled | No se anade |

#### Funciones helper
- `ParseString(m, key)` - Extrae string de un mapa
- `ParseBool(m, key)` - Extrae boolean (soporta "1"/"0", "true"/"false")
- `ParseInt(m, key)` - Extrae int64 de string o float64
- `GetEntryContent(resp)` - Navega la estructura `feed > entry[0] > content` del REST API
- `GetEntryACL(resp)` - Navega la estructura `feed > entry[0] > acl`
- `URLEncode(s)` - Wrapper sobre `url.PathEscape` para segmentos de URL

### 2.2 Endpoints API

#### API REST Tradicional (`/servicesNS/{owner}/{app}/...`)
| Modulo | Endpoint Base |
|--------|---------------|
| `saved_searches.go` | `/servicesNS/{owner}/{app}/saved/searches` |
| `macros.go` | `/servicesNS/{owner}/{app}/configs/conf-macros` |
| `lookups.go` | `/servicesNS/{owner}/{app}/data/transforms/lookups` y `/data/lookup-table-files` |
| `kvstore.go` | `/servicesNS/{owner}/{app}/storage/collections/config` |
| `analytic_stories.go` | `/servicesNS/{owner}/{app}/configs/conf-analyticstories` |
| `threat_intel.go` | `/services/data/threat_intel/item/{collection}` |

#### API ES v2 (`/servicesNS/nobody/missioncontrol/public/v2/...`)
| Modulo | Endpoint Base |
|--------|---------------|
| `investigations.go` | `/servicesNS/nobody/missioncontrol/public/v2/investigations` |
| `findings.go` | `/servicesNS/nobody/missioncontrol/public/v2/findings` |
| `risks.go` | `/servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores` |
| `assets_identity.go` | `/servicesNS/nobody/missioncontrol/public/v2/assets` y `/identities` |

---

## 3. Recursos Terraform (12)

### 3.1 Recursos basados en REST API (7)

Todos siguen el patron CRUD estandar de Splunk REST:
- **Create:** `POST` al endpoint de coleccion (sin nombre en la ruta)
- **Read:** `GET` al endpoint especifico (con nombre en la ruta)
- **Update:** `POST` al endpoint especifico (Splunk usa POST, no PUT/PATCH)
- **Delete:** `DELETE` al endpoint especifico

#### `splunkes_correlation_search` (847 lineas)
Recurso mas complejo. Gestiona correlation searches de ES como saved searches con parametros adicionales:
- **Campos core:** `name`, `search`, `cron_schedule`, `description`, `disabled`
- **Parametros ES:** `correlation_search_enabled`, `security_domain`, `schedule_priority`
- **Accion Notable:** `notable_enabled`, `notable_severity`, `notable_rule_title`, `notable_security_domain`, `notable_drilldown_name/search`, `notable_default_owner`, `notable_recommended_actions`
- **Accion Risk:** `risk_enabled`, `risk_score`, `risk_message`, `risk_object_field`, `risk_object_type`
- **MITRE ATT&CK:** `mitre_attack_ids` (lista), `kill_chain_phases`, `cis20`, `nist`, `analytic_story` - almacenados como JSON en `action.correlationsearch.annotations`
- **Alertas:** `alert_type`, `alert_comparator`, `alert_threshold`, `alert_suppress`, `alert_suppress_period`, `alert_suppress_fields`

#### `splunkes_saved_search` (415 lineas)
Saved searches genericos con soporte para alertas y scheduling. App por defecto: `"search"`.

#### `splunkes_macro` (330 lineas)
Macros de busqueda SPL con soporte para argumentos, validacion e iseval.

#### `splunkes_lookup_definition` (383 lineas)
Definiciones de lookup que pueden ser respaldadas por CSV (`filename`) o KV Store (`collection` + `external_type`).

#### `splunkes_lookup_table` (273 lineas)
Gestion de archivos CSV de lookup, incluyendo carga de contenido via el endpoint `lookup_edit/lookup_contents`.

#### `splunkes_kvstore_collection` (313 lineas)
Colecciones KV Store con esquemas de campos (`fields` como `types.Map`) y campos acelerados.

#### `splunkes_analytic_story` (359 lineas)
Agrupacion de detecciones con listas de `detection_searches`, `investigative_searches`, `contextual_searches`.

### 3.2 Recursos basados en ES v2 API (3)

#### `splunkes_investigation` (392 lineas)
Ciclo de vida de investigaciones. **Delete cierra la investigacion** (no hay endpoint DELETE en ES v2). Mapeos: `name` -> `title`, `priority` -> `urgency`.

#### `splunkes_investigation_note` (325 lineas)
Notas en investigaciones. **Import ID compuesto:** `investigation_id/note_id`.

#### `splunkes_finding` (368 lineas)
Hallazgos de seguridad. **No se puede actualizar ni eliminar via API.** Update persiste el estado con warning. Delete elimina del estado de Terraform con warning.

### 3.3 Recursos hibridos (2)

#### `splunkes_risk_modifier` (290 lineas)
Modificadores de riesgo para entidades. **Import ID compuesto:** `entity/entity_type`. Delete solo elimina del estado de Terraform.

#### `splunkes_threat_intel` (300 lineas)
Items de inteligencia de amenazas. **Import ID compuesto:** `collection/key`. Usa REST API tradicional pero con estructura diferente.

### 3.4 Grafo de Dependencias

```
splunkes_lookup_table ─────────────────────────┐
                                               ▼
splunkes_kvstore_collection ──► splunkes_lookup_definition ──┐
                                                             │
splunkes_macro ──────────────────────────────────────────────┤
                                                             │
splunkes_threat_intel ──────────────────────────────────────┤
                                                             │
splunkes_analytic_story ────────────────────────────────────┤
                                                             ▼
                                        splunkes_correlation_search
                                                             │
                                                             ▼
                                        splunkes_finding ─► splunkes_investigation
                                                                        │
                                        splunkes_risk_modifier          ▼
                                                        splunkes_investigation_note
```

Las dependencias se manejan **implicitamente** via referencias Terraform (ej: `splunkes_macro.my_macro.name` en una query SPL) o **explicitamente** via bloques `depends_on`.

---

## 4. Data Sources (7)

Todos son de solo lectura y siguen el patron estandar:
- Interface assertions: `datasource.DataSource` + `datasource.DataSourceWithConfigure`
- Valores por defecto aplicados programaticamente en `Read()` (no via campo `Default` del schema)

| Data Source | API | Campo requerido | Campos computados |
|-------------|-----|-----------------|-------------------|
| `splunkes_correlation_search` | REST Saved Searches | `name` | search, description, disabled, cron_schedule, notable_enabled, security_domain, severity, risk_score |
| `splunkes_investigation` | ES v2 | `id` | name, status, assignee, priority |
| `splunkes_finding` | ES v2 | `id` | rule_title, security_domain, risk_score, risk_object, risk_object_type, severity |
| `splunkes_risk_score` | ES v2 | `entity` | entity_type, risk_score, risk_level |
| `splunkes_identity` | ES v2 | `id` | first_name, last_name, email, bunit, category, priority, watchlist |
| `splunkes_asset` | ES v2 | `id` | ip, mac, dns, nt_host, bunit, category, priority, is_expected |
| `splunkes_macro` | REST Macros | `name` | app, owner, definition, description, args, validation, iseval |

---

## 5. Problemas Encontrados y Soluciones

### 5.1 Incompatibilidad de Version Go

**Problema:** `terraform-plugin-framework v1.16.0` requiere `go >= 1.24.0`, pero el toolchain local es `go1.22.1`.

**Solucion:** Downgrade a `terraform-plugin-framework v1.13.0` y configuracion de `go 1.22.0` en `go.mod`.

### 5.2 Permisos de Cache de Go

**Problema:** `go build` fallaba con "permission denied" en `/Users/cx02875/go/pkg/mod/cache/`.

**Solucion:** Variables de entorno `GOMODCACHE=/tmp/gomodcache GOCACHE=/tmp/gobuildcache`.

### 5.3 Campo `Default` en Data Sources

**Problema:** `datasource/schema.StringAttribute` no soporta el campo `Default` (a diferencia de `resource/schema`). Causaba errores de compilacion.

**Solucion:** Eliminacion de `Default`/`stringdefault` en data sources. Valores por defecto aplicados manualmente en la funcion `Read()`:
```go
if state.App.IsNull() || state.App.IsUnknown() {
    state.App = types.StringValue("SplunkEnterpriseSecuritySuite")
}
```

### 5.4 ImportState con `path.Root()`

**Problema:** `resource_correlation_search.go` usaba strings literales (`"id"`) en vez de `path.Root("id")` para `ImportStatePassthroughID` y `SetAttribute`.

**Solucion:** Anadido import de `path` y conversion de todas las llamadas a `path.Root()`.

### 5.5 ImportState Incompleto en 8 Recursos

**Problema:** Despues de `ImportState`, la funcion `Read()` necesita `name`, `owner`, `app` (o `collection`/`key`) del estado, pero solo `id` estaba configurado. Esto causaba que `terraform import` fallara para todos los recursos REST.

**Recursos afectados:**
- `resource_saved_search.go`
- `resource_macro.go`
- `resource_lookup_definition.go`
- `resource_lookup_table.go`
- `resource_kvstore_collection.go`
- `resource_analytic_story.go`
- `resource_threat_intel.go` (composite ID: `collection/key`)

**Solucion:** Se anadieron llamadas `SetAttribute` en `ImportState` para cada recurso:
```go
func (r *MacroResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}
```

Para `resource_threat_intel.go` se implemento parsing de ID compuesto:
```go
parts := strings.SplitN(req.ID, "/", 2)
resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("collection"), parts[0])...)
```

### 5.6 Finding Update No Persistia Estado

**Problema:** `resource_finding.go` Update anadida un warning pero no llamaba a `resp.State.Set()`, causando diffs perpetuos en Terraform.

**Solucion:** Se lee el estado planificado y se persiste:
```go
func (r *findingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    resp.Diagnostics.AddWarning("Findings cannot be updated", "...")
    var state findingResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

### 5.7 Body Exhaustion en Reintentos HTTP

**Problema:** En `DoRequest` y `DoJSONRequest`, el `io.Reader` para el body se creaba una sola vez antes del loop de reintentos. Despues del primer intento, el reader estaba consumido y los reintentos enviaban un body vacio.

**Solucion:** El `io.Reader` ahora se crea **dentro** de cada iteracion del loop de reintentos:
```go
// DoRequest - antes (bug):
reqBody = strings.NewReader(body.Encode())
for attempt := 0; attempt < 5; attempt++ { ... }

// DoRequest - despues (corregido):
encodedBody = body.Encode()
for attempt := 0; attempt < 5; attempt++ {
    var reqBody io.Reader
    if encodedBody != "" {
        reqBody = strings.NewReader(encodedBody)
    }
    ...
}
```

Aplicado analogamente a `DoJSONRequest` con `bytes.NewReader(jsonData)`.

---

## 6. Proceso de Revision de Codigo

Se ejecutaron **3 agentes de revision en paralelo** para auditar todo el codigo:

### Revision 1: Correlation Search + Saved Search
- Verifico endpoints, parametros, metodos HTTP
- **Encontro:** SavedSearch ImportState no configura name/owner/app
- **Encontro:** SavedSearch Read no maneja 404

### Revision 2: Recursos ES v2
- Verifico uso de DoJSONRequest vs DoRequest
- Verifico semantica de recursos no-eliminables
- **Encontro:** Finding Update no persiste estado
- **Encontro:** Body exhaustion en reintentos HTTP
- **Encontro:** Risk modifier composite ID podria fallar si entity contiene "/"

### Revision 3: Macro/Lookup/KVStore/ThreatIntel/AnalyticStory
- Verifico endpoints, URL encoding, dependencias
- **Encontro:** ImportState incompleto en 6 recursos
- **Encontro:** ThreatIntel necesita parsing de composite ID

### Agente de Correccion
Un agente adicional corrigio los 8 archivos afectados y verifico que la compilacion pasaba limpiamente.

---

## 7. Configuracion del Proveedor

### Variables de Entorno Soportadas

| Variable | Descripcion |
|----------|-------------|
| `SPLUNK_URL` | URL base del servidor Splunk (ej: `https://splunk:8089`) |
| `SPLUNK_USERNAME` | Usuario para autenticacion basica |
| `SPLUNK_PASSWORD` | Contrasena para autenticacion basica |
| `SPLUNK_AUTH_TOKEN` | Token de autenticacion Bearer |
| `SPLUNK_INSECURE_SKIP_VERIFY` | Omitir verificacion TLS (`true`/`false`) |
| `SPLUNK_TIMEOUT` | Timeout de peticiones HTTP en segundos |

### Configuracion HCL

```hcl
provider "splunkes" {
  url                  = "https://splunk.example.com:8089"
  auth_token           = var.splunk_token
  insecure_skip_verify = true
  timeout              = 30
}
```

---

## 8. Build y Release

### Compilacion Local
```bash
GOMODCACHE=/tmp/gomodcache GOCACHE=/tmp/gobuildcache go build ./...
```

### Makefile Targets
```bash
make build      # Compilar el proveedor
make test       # Tests unitarios
make testacc    # Tests de aceptacion (requiere Splunk)
make fmt        # Formatear codigo Go
make vet        # Analisis estatico
make lint       # Linting completo
```

### GoReleaser
Configurado en `.goreleaser.yml` para release multiplataforma:
- linux/amd64, linux/arm64
- darwin/amd64, darwin/arm64
- windows/amd64

---

## 9. Proximos Pasos

1. **Inicializar repositorio Git** y hacer commit inicial
2. **Tests de aceptacion** contra una instancia real de Splunk ES (`make testacc`)
3. **Publicar en Terraform Registry** via GoReleaser + firma GPG
4. **Integrar con Terragrunt** para patrones DRY multi-entorno (dev/staging/prod)
5. **Pipeline CI/CD** para validacion automatica de cambios

---

## 10. Metricas del Proyecto

| Metrica | Valor |
|---------|-------|
| Archivos Go | 32 |
| Lineas de codigo Go | 6.703 |
| Recursos Terraform | 12 |
| Data Sources | 7 |
| Modulos SDK | 11 |
| Archivos de ejemplo | 9 |
| Tamano del binario | 23 MB |
| Bugs encontrados y corregidos | 7 |
| Agentes de revision ejecutados | 4 |
