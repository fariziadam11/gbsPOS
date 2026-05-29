# GBS POS-CMS API - Architecture Diagram

## 1. System Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        Android[📱 Android POS App<br/>Sunmi D3 Pro]
        Display[🖥️ Customer Display<br/>Secondary Screen]
        Printer[🖨️ USB Thermal Printer]
        Browser[🌐 Web Browser<br/>CMS Admin Panel]
    end
    
    subgraph "API Gateway / Load Balancer"
        LB[⚖️ Load Balancer<br/>Optional: Nginx/Traefik]
    end
    
    subgraph "Backend Services - Docker Compose"
        subgraph "POS API Service :8080"
            POS[🏪 POS API<br/>Golang + Gin]
            POSAuth[🔐 JWT Auth]
            POSMiddleware[🛡️ Middleware<br/>CORS, Logger, Auth]
        end
        
        subgraph "CMS API Service :8081"
            CMS[📺 CMS API<br/>Golang + Gin]
            CMSAuth[🔐 JWT Auth]
            CMSMiddleware[🛡️ Middleware<br/>CORS, Logger, Auth]
            FileStorage[📁 File Storage<br/>/uploads/ads]
        end
        
        subgraph "Shared Module"
            Common[📦 gbs-common<br/>Shared Utilities]
        end
    end
    
    subgraph "Database Layer"
        DB[(🗄️ PostgreSQL 15<br/>Database: gbs_pos<br/>Port: 5433)]
    end
    
    subgraph "External Services"
        Neurogine[💳 Neurogine SoftPOS<br/>Card Payment Gateway]
        QRIS[📱 QRIS Payment<br/>QR Code Payment]
    end
    
    Android -->|HTTP/REST| LB
    Browser -->|HTTP/REST| LB
    
    LB -->|Route /v1/*| POS
    LB -->|Route /v1/*| CMS
    
    Android -->|Display Cart| Display
    Android -->|Print Receipt| Printer
    Android -->|Process Card| Neurogine
    Android -->|Process QRIS| QRIS
    
    POS --> POSAuth
    POS --> POSMiddleware
    CMS --> CMSAuth
    CMS --> CMSMiddleware
    
    POSMiddleware --> Common
    CMSMiddleware --> Common
    
    POS -->|Read/Write| DB
    CMS -->|Read/Write| DB
    CMS -->|Store Videos| FileStorage
    
    Neurogine -.->|Transaction Data| Android
    QRIS -.->|Payment Status| Android
    
    style Android fill:#95e1d3,stroke:#38ada9,stroke-width:3px
    style POS fill:#4ecdc4,stroke:#0a8a7a,stroke-width:3px
    style CMS fill:#ff6b6b,stroke:#c92a2a,stroke-width:3px
    style DB fill:#ffd93d,stroke:#f39c12,stroke-width:3px
    style Common fill:#a29bfe,stroke:#6c5ce7,stroke-width:2px
```

## 2. Detailed Backend Architecture

```mermaid
graph TB
    subgraph "POS API Architecture :8080"
        subgraph "Entry Point"
            POSMain[main.go<br/>Server Entry Point]
        end
        
        subgraph "HTTP Layer"
            POSRouter[Gin Router<br/>Route Configuration]
            POSHandler[Handlers<br/>- AuthHandler<br/>- ProductHandler<br/>- OrderHandler<br/>- SettlementHandler]
        end
        
        subgraph "Business Logic Layer"
            POSService[Services<br/>- AuthService<br/>- ProductService<br/>- OrderService<br/>- SettlementService]
        end
        
        subgraph "Data Access Layer"
            POSRepo[Repositories<br/>- UserRepo<br/>- ProductRepo<br/>- OrderRepo<br/>- SettlementRepo]
        end
        
        subgraph "Domain Layer"
            POSModel[Models<br/>- User<br/>- Product<br/>- Order<br/>- Settlement]
        end
        
        POSMain --> POSRouter
        POSRouter --> POSHandler
        POSHandler --> POSService
        POSService --> POSRepo
        POSRepo --> POSModel
    end
    
    subgraph "CMS API Architecture :8081"
        subgraph "Entry Point"
            CMSMain[main.go<br/>Server Entry Point]
        end
        
        subgraph "HTTP Layer"
            CMSRouter[Gin Router<br/>Route Configuration]
            CMSHandler[Handlers<br/>- AuthHandler<br/>- AdHandler<br/>- PlaylistHandler]
        end
        
        subgraph "Business Logic Layer"
            CMSService[Services<br/>- AuthService<br/>- AdService<br/>- PlaylistService]
        end
        
        subgraph "Data Access Layer"
            CMSRepo[Repositories<br/>- UserRepo<br/>- AdRepo<br/>- PlaylistRepo]
        end
        
        subgraph "Domain Layer"
            CMSModel[Models<br/>- User<br/>- Ad<br/>- Playlist<br/>- AdPlayLog]
        end
        
        CMSMain --> CMSRouter
        CMSRouter --> CMSHandler
        CMSHandler --> CMSService
        CMSService --> CMSRepo
        CMSRepo --> CMSModel
    end
    
    subgraph "Shared Common Module"
        CommonMiddleware[Middleware<br/>- AuthMiddleware<br/>- CORSMiddleware<br/>- LoggerMiddleware]
        CommonPkg[Packages<br/>- Response<br/>- Utils]
    end
    
    POSRouter -.->|Import| CommonMiddleware
    CMSRouter -.->|Import| CommonMiddleware
    POSHandler -.->|Use| CommonPkg
    CMSHandler -.->|Use| CommonPkg
    
    subgraph "Database"
        POSDB[(PostgreSQL<br/>POS Tables)]
        CMSDB[(PostgreSQL<br/>CMS Tables)]
    end
    
    POSRepo --> POSDB
    CMSRepo --> CMSDB
    
    style POSMain fill:#4ecdc4
    style CMSMain fill:#ff6b6b
    style CommonMiddleware fill:#a29bfe
    style POSDB fill:#ffd93d
    style CMSDB fill:#ffd93d
```

## 3. Database Schema Architecture

```mermaid
erDiagram
    USERS ||--o{ ORDERS : creates
    PRODUCTS ||--o{ ORDER_ITEMS : contains
    ORDERS ||--|{ ORDER_ITEMS : has
    ORDERS }o--|| SETTLEMENTS : included_in
    ADS ||--o{ PLAYLIST_ADS : included_in
    PLAYLISTS ||--o{ PLAYLIST_ADS : contains
    ADS ||--o{ AD_PLAY_LOGS : tracks
    
    USERS {
        int id PK
        string username UK
        string password_hash
        string name
        string role
        timestamp created_at
        timestamp updated_at
    }
    
    PRODUCTS {
        int id PK
        string name
        decimal price
        string category
        string image_url
        string store_type
        timestamp created_at
        timestamp updated_at
    }
    
    ORDERS {
        string id PK
        decimal subtotal
        decimal tax
        decimal total
        string payment_method
        decimal cash_received
        decimal change_amount
        bigint timestamp
        boolean is_voided
        boolean is_settled
        string transaction_id
        string approval_code
        string store_type
        string terminal_id
        timestamp created_at
    }
    
    ORDER_ITEMS {
        int id PK
        string order_id FK
        int product_id
        string product_name
        decimal product_price
        int qty
        decimal subtotal
    }
    
    SETTLEMENTS {
        string id PK
        bigint timestamp
        int batch_count
        decimal total_amount
        decimal card_total
        decimal qris_total
        decimal cash_total
        string status
        string store_type
        string terminal_id
        timestamp created_at
    }
    
    ADS {
        int id PK
        string title
        string file_path
        bigint file_size
        int duration
        json store_types
        date start_date
        date end_date
        boolean is_active
        int play_count
        timestamp created_at
    }
    
    PLAYLISTS {
        int id PK
        string name
        string description
        boolean is_active
        timestamp created_at
    }
    
    PLAYLIST_ADS {
        int id PK
        int playlist_id FK
        int ad_id FK
        int play_order
        timestamp created_at
    }
    
    AD_PLAY_LOGS {
        int id PK
        int ad_id FK
        string terminal_id
        timestamp played_at
    }
```

## 4. Request Flow Architecture

```mermaid
sequenceDiagram
    participant Client as 📱 Android App
    participant LB as ⚖️ Load Balancer
    participant Auth as 🔐 Auth Middleware
    participant Handler as 🎯 Handler
    participant Service as 💼 Service
    participant Repo as 📊 Repository
    participant DB as 🗄️ PostgreSQL
    participant Cache as 💾 Local Cache
    
    Note over Client,DB: Authentication Flow
    Client->>LB: POST /v1/login
    LB->>Handler: Route Request
    Handler->>Service: Validate Credentials
    Service->>Repo: Find User
    Repo->>DB: SELECT * FROM users
    DB-->>Repo: User Data
    Repo-->>Service: User Object
    Service->>Service: Verify Password
    Service->>Service: Generate JWT Token
    Service-->>Handler: Token + User Info
    Handler-->>Client: 200 OK + JWT
    
    Note over Client,DB: Authenticated Request Flow
    Client->>LB: GET /v1/products<br/>Authorization: Bearer <token>
    LB->>Auth: Verify Token
    Auth->>Auth: Validate JWT
    Auth->>Auth: Extract Claims
    Auth->>Handler: Request + User Context
    Handler->>Service: Get Products
    Service->>Repo: Find All Products
    Repo->>DB: SELECT * FROM products
    DB-->>Repo: Product List
    Repo-->>Service: Products
    Service-->>Handler: Products
    Handler-->>Client: 200 OK + Products
    
    Note over Client,DB: Offline-First Sync Flow
    Client->>Cache: Save Order Locally
    Cache-->>Client: Saved
    Client->>LB: POST /v1/orders (Retry)
    LB->>Auth: Verify Token
    Auth->>Handler: Create Order
    Handler->>Service: Process Order
    Service->>Repo: Check Duplicate (Idempotent)
    Repo->>DB: SELECT * FROM orders WHERE id = ?
    alt Order Exists
        DB-->>Repo: Existing Order
        Repo-->>Service: Order Found
        Service-->>Handler: Existing Order
        Handler-->>Client: 200 OK (idempotent: true)
    else New Order
        DB-->>Repo: Not Found
        Repo->>DB: INSERT INTO orders
        DB-->>Repo: Created
        Repo-->>Service: New Order
        Service-->>Handler: New Order
        Handler-->>Client: 201 Created
    end
```

## 5. Deployment Architecture

```mermaid
graph TB
    subgraph "Production Environment"
        subgraph "Docker Host Server"
            subgraph "Docker Compose Stack"
                DC[🐳 Docker Compose<br/>Orchestration]
                
                subgraph "Container 1"
                    POS_C[POS API Container<br/>Image: gbs-pos-api<br/>Port: 8080]
                end
                
                subgraph "Container 2"
                    CMS_C[CMS API Container<br/>Image: gbs-cms-api<br/>Port: 8081]
                end
                
                subgraph "Container 3"
                    DB_C[PostgreSQL Container<br/>Image: postgres:15-alpine<br/>Port: 5432]
                end
            end
            
            subgraph "Docker Volumes"
                VOL1[📦 pgdata<br/>Database Storage]
                VOL2[📦 pos-uploads<br/>POS Files]
                VOL3[📦 cms-uploads<br/>Video Files]
            end
        end
        
        subgraph "Network"
            NET[🌐 Docker Bridge Network<br/>Internal Communication]
        end
    end
    
    subgraph "External Access"
        Internet[🌍 Internet]
        Firewall[🔥 Firewall<br/>Port 8080, 8081]
    end
    
    DC --> POS_C
    DC --> CMS_C
    DC --> DB_C
    
    POS_C --> VOL2
    CMS_C --> VOL3
    DB_C --> VOL1
    
    POS_C -.->|Internal| NET
    CMS_C -.->|Internal| NET
    DB_C -.->|Internal| NET
    
    Internet --> Firewall
    Firewall --> POS_C
    Firewall --> CMS_C
    
    style POS_C fill:#4ecdc4
    style CMS_C fill:#ff6b6b
    style DB_C fill:#ffd93d
    style VOL1 fill:#a29bfe
    style VOL2 fill:#a29bfe
    style VOL3 fill:#a29bfe
```

## 6. Multi-Store Architecture

```mermaid
graph TB
    subgraph "Store Types"
        RETAIL[🏪 RETAIL Store<br/>Snacks, Beverages<br/>Household, Personal Care]
        FNB[🍔 F&B Store<br/>Food, Beverages<br/>Desserts]
        OUTFIT[👔 OUTFIT Store<br/>Tops, Bottoms<br/>Outerwear, Accessories]
    end
    
    subgraph "POS Terminals"
        T1[📱 Terminal POS-001<br/>RETAIL]
        T2[📱 Terminal POS-002<br/>F&B]
        T3[📱 Terminal POS-003<br/>OUTFIT]
        T4[📱 Terminal POS-004<br/>RETAIL]
    end
    
    subgraph "Backend API"
        API[🔄 POS API<br/>Multi-Store Support]
    end
    
    subgraph "Database - Shared"
        subgraph "Products Table"
            P1[Products<br/>store_type = 'RETAIL']
            P2[Products<br/>store_type = 'FNB']
            P3[Products<br/>store_type = 'OUTFIT']
        end
        
        subgraph "Orders Table"
            O1[Orders<br/>store_type + terminal_id]
        end
        
        subgraph "Settlements Table"
            S1[Settlements<br/>Grouped by store_type<br/>and terminal_id]
        end
    end
    
    RETAIL --> T1 & T4
    FNB --> T2
    OUTFIT --> T3
    
    T1 & T2 & T3 & T4 -->|HTTP/REST| API
    
    API --> P1 & P2 & P3
    API --> O1
    API --> S1
    
    style RETAIL fill:#51cf66
    style FNB fill:#ff6b6b
    style OUTFIT fill:#339af0
    style API fill:#4ecdc4
```

## 7. Security Architecture

```mermaid
graph TB
    subgraph "Security Layers"
        subgraph "Layer 1: Network Security"
            FW[🔥 Firewall<br/>Port Restrictions]
            HTTPS[🔒 HTTPS/TLS<br/>SSL Certificates]
        end
        
        subgraph "Layer 2: Authentication"
            JWT[🎫 JWT Token<br/>HS256 Algorithm]
            Login[🔐 Login Endpoint<br/>Username + Password]
            BCrypt[🔑 BCrypt Hashing<br/>Password Storage]
        end
        
        subgraph "Layer 3: Authorization"
            RBAC[👥 Role-Based Access<br/>ADMIN vs CASHIER]
            Middleware[🛡️ Auth Middleware<br/>Token Verification]
        end
        
        subgraph "Layer 4: Data Security"
            SQL[💉 SQL Injection Prevention<br/>GORM Parameterized Queries]
            Validation[✅ Input Validation<br/>go-playground/validator]
            CORS[🌐 CORS Policy<br/>Allowed Origins]
        end
        
        subgraph "Layer 5: Audit & Logging"
            Logger[📝 Request Logger<br/>zerolog]
            Audit[📊 Audit Trail<br/>created_at, updated_at]
        end
    end
    
    Client[📱 Client Request]
    
    Client --> FW
    FW --> HTTPS
    HTTPS --> Login
    Login --> BCrypt
    BCrypt --> JWT
    JWT --> Middleware
    Middleware --> RBAC
    RBAC --> Validation
    Validation --> SQL
    SQL --> CORS
    CORS --> Logger
    Logger --> Audit
    
    style FW fill:#ff6b6b
    style JWT fill:#4ecdc4
    style RBAC fill:#ffd93d
    style SQL fill:#51cf66
    style Logger fill:#a29bfe
```

## 8. Payment Processing Architecture

```mermaid
graph TB
    subgraph "Payment Methods"
        Cash[💵 Cash Payment]
        Card[💳 Card Payment]
        QRIS[📱 QRIS Payment]
    end
    
    subgraph "Android POS App"
        Cart[🛒 Shopping Cart]
        Checkout[💰 Checkout Screen]
        PaymentUI[💳 Payment UI]
    end
    
    subgraph "External Payment Gateways"
        Neurogine[💳 Neurogine SoftPOS SDK<br/>Card Processing]
        QRISGateway[📱 QRIS Gateway<br/>QR Code Processing]
    end
    
    subgraph "Backend Processing"
        OrderAPI[📦 Order API<br/>POST /v1/orders]
        Validation[✅ Validation<br/>- Amount Check<br/>- Payment Method<br/>- Idempotency]
        Storage[(💾 Database<br/>Order Storage)]
    end
    
    subgraph "Receipt & Display"
        Printer[🖨️ Thermal Printer<br/>Receipt Print]
        Display[🖥️ Customer Display<br/>Transaction Summary]
    end
    
    Cart --> Checkout
    Checkout --> PaymentUI
    
    PaymentUI -->|Select| Cash
    PaymentUI -->|Select| Card
    PaymentUI -->|Select| QRIS
    
    Cash -->|Enter Amount| OrderAPI
    
    Card -->|Process| Neurogine
    Neurogine -->|Transaction Data<br/>- Transaction ID<br/>- Approval Code<br/>- Masked Account| OrderAPI
    
    QRIS -->|Scan & Pay| QRISGateway
    QRISGateway -->|Payment Status| OrderAPI
    
    OrderAPI --> Validation
    Validation -->|Valid| Storage
    Storage -->|Success| Printer
    Storage -->|Success| Display
    
    style Cash fill:#51cf66
    style Card fill:#339af0
    style QRIS fill:#ff6b6b
    style OrderAPI fill:#4ecdc4
    style Storage fill:#ffd93d
```

## 9. Settlement Process Architecture

```mermaid
flowchart TB
    Start([🎯 Admin Initiates Settlement])
    
    subgraph "Pre-Settlement"
        Check[📊 Check Unsettled Orders<br/>GET /orders/unsettled/summary]
        Display[📈 Display Summary<br/>- Total Orders<br/>- Total Amount<br/>- By Payment Method]
        Confirm{✅ Confirm Settlement?}
    end
    
    subgraph "Settlement Transaction"
        Lock[🔒 Database Transaction<br/>BEGIN TRANSACTION]
        Select[🔍 SELECT FOR UPDATE<br/>Lock Unsettled Orders]
        Filter[🎯 Filter Orders<br/>- is_settled = false<br/>- is_voided = false<br/>- By store_type<br/>- By terminal_id]
        Calculate[🧮 Calculate Totals<br/>- Count Orders<br/>- Sum by Payment Method<br/>- Card Total<br/>- QRIS Total<br/>- Cash Total]
        Create[📝 Create Settlement Record<br/>ID: SETTLE-timestamp]
        Update[✏️ Update Orders<br/>SET is_settled = true]
        Commit[✅ COMMIT TRANSACTION]
    end
    
    subgraph "Post-Settlement"
        Report[📄 Generate Report<br/>Settlement Summary]
        Print[🖨️ Print Settlement Report]
        Notify[📢 Notify Admin<br/>Settlement Complete]
    end
    
    Error[❌ Rollback Transaction]
    
    Start --> Check
    Check --> Display
    Display --> Confirm
    
    Confirm -->|Yes| Lock
    Confirm -->|No| Cancel([❌ Cancel])
    
    Lock --> Select
    Select --> Filter
    Filter --> Calculate
    Calculate --> Create
    Create --> Update
    Update --> Success{Success?}
    
    Success -->|Yes| Commit
    Success -->|No| Error
    
    Commit --> Report
    Report --> Print
    Print --> Notify
    Notify --> End([✅ Settlement Complete])
    
    Error --> End2([❌ Settlement Failed])
    
    style Start fill:#4ecdc4
    style Lock fill:#ff6b6b
    style Calculate fill:#ffd93d
    style Commit fill:#51cf66
    style Error fill:#e74c3c
```

## 10. CMS Video Streaming Architecture

```mermaid
graph TB
    subgraph "Client Side"
        Browser[🌐 Web Browser<br/>Admin Panel]
        Player[📺 Video Player<br/>ExoPlayer/HTML5]
    end
    
    subgraph "CMS API :8081"
        Upload[📤 Upload Endpoint<br/>POST /ads/upload]
        Download[📥 Download Endpoint<br/>GET /ads/:id/download]
        Stream[🎬 Streaming Handler<br/>Range Request Support]
    end
    
    subgraph "File Storage"
        FS[📁 File System<br/>/uploads/ads/]
        Video1[🎥 video1.mp4]
        Video2[🎥 video2.mp4]
        Video3[🎥 video3.mp4]
    end
    
    subgraph "Database"
        AdTable[(📊 Ads Table<br/>- file_path<br/>- file_size<br/>- duration<br/>- store_types)]
    end
    
    subgraph "Video Processing"
        Validation[✅ Validation<br/>- Max 50MB<br/>- Video Format<br/>- Duration Check]
        Metadata[📋 Extract Metadata<br/>- File Size<br/>- Duration<br/>- Format]
    end
    
    Browser -->|Upload Video| Upload
    Upload --> Validation
    Validation -->|Valid| Metadata
    Metadata --> FS
    FS --> Video1 & Video2 & Video3
    Metadata --> AdTable
    
    Player -->|Request Video| Download
    Download --> Stream
    Stream -->|Range: bytes=0-1024| FS
    FS -->|Partial Content 206| Stream
    Stream -->|Video Chunks| Player
    
    AdTable -.->|Metadata| Download
    
    style Upload fill:#51cf66
    style Stream fill:#339af0
    style FS fill:#ffd93d
    style Player fill:#ff6b6b
```

## 11. Monitoring & Logging Architecture

```mermaid
graph TB
    subgraph "Application Layer"
        POS[POS API]
        CMS[CMS API]
    end
    
    subgraph "Logging System"
        Logger[📝 Zerolog Logger<br/>Structured Logging]
        
        subgraph "Log Levels"
            Debug[🔍 DEBUG<br/>Development Info]
            Info[ℹ️ INFO<br/>General Info]
            Warn[⚠️ WARN<br/>Warnings]
            Error[❌ ERROR<br/>Errors]
        end
    end
    
    subgraph "Log Output"
        Console[🖥️ Console Output<br/>Docker Logs]
        File[📄 Log Files<br/>Optional: File Storage]
    end
    
    subgraph "Monitoring Metrics"
        Request[📊 Request Metrics<br/>- Method<br/>- Path<br/>- Status Code<br/>- Duration]
        Database[💾 Database Metrics<br/>- Query Time<br/>- Connection Pool]
        System[⚙️ System Metrics<br/>- CPU Usage<br/>- Memory Usage]
    end
    
    subgraph "Health Checks"
        Health[❤️ Health Endpoint<br/>GET /health]
        DBHealth[💓 Database Health<br/>Connection Check]
    end
    
    POS --> Logger
    CMS --> Logger
    
    Logger --> Debug & Info & Warn & Error
    
    Debug & Info & Warn & Error --> Console
    Debug & Info & Warn & Error --> File
    
    POS --> Request
    CMS --> Request
    
    POS --> Database
    CMS --> Database
    
    POS --> System
    CMS --> System
    
    POS --> Health
    CMS --> Health
    Health --> DBHealth
    
    style Logger fill:#a29bfe
    style Console fill:#4ecdc4
    style Request fill:#ffd93d
    style Health fill:#51cf66
```

## 12. Technology Stack Overview

```mermaid
mindmap
  root((GBS POS-CMS<br/>Technology Stack))
    Backend
      Language
        Go 1.26
      Framework
        Gin Web Framework
      ORM
        GORM
      Migration
        golang-migrate
      Auth
        JWT golang-jwt/jwt
      Validation
        go-playground/validator
      Logging
        zerolog
      Config
        caarlos0/env
    Database
      PostgreSQL 15
      GORM Driver
      Connection Pool
      Migrations
    Frontend
      Android
        Kotlin
        Jetpack Compose
        Room Database
        Retrofit
        Coroutines
      Web
        React Optional
        HTML5 Video
    Infrastructure
      Docker
      Docker Compose
      Linux Alpine
    External Services
      Neurogine SoftPOS
      QRIS Payment
      Thermal Printer
    Security
      JWT HS256
      BCrypt
      CORS
      Input Validation
      SQL Injection Prevention
```

---

## Summary

Dokumentasi ini mencakup:

1. ✅ **System Architecture Overview** - Gambaran umum sistem
2. ✅ **Detailed Backend Architecture** - Arsitektur backend detail
3. ✅ **Database Schema Architecture** - Skema database ER Diagram
4. ✅ **Request Flow Architecture** - Alur request sequence diagram
5. ✅ **Deployment Architecture** - Arsitektur deployment Docker
6. ✅ **Multi-Store Architecture** - Arsitektur multi-toko
7. ✅ **Security Architecture** - Arsitektur keamanan berlapis
8. ✅ **Payment Processing Architecture** - Arsitektur pemrosesan pembayaran
9. ✅ **Settlement Process Architecture** - Proses settlement detail
10. ✅ **CMS Video Streaming Architecture** - Arsitektur streaming video
11. ✅ **Monitoring & Logging Architecture** - Arsitektur monitoring dan logging
12. ✅ **Technology Stack Overview** - Ringkasan teknologi yang digunakan

Semua diagram dapat di-render menggunakan:
- GitHub (otomatis)
- VS Code + Markdown Preview Mermaid Support
- https://mermaid.live
- Obsidian
- Notion (dengan plugin)

**Siap untuk presentasi kepada atasan!** 🎯
