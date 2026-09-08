# Go Cookbook 项目

一个完整的菜谱管理系统，包含后端API服务和前端管理界面，用于管理菜品、食材和相关产品信息。并包含二维码生成器，用于生成跳转网页的二维码，用户扫描二维码后可以直接跳转到详情页面。适用于超市的食品商品介绍和推广服务。

## 项目结构

```
go-cookbook/
├── backend/              # 后端Go语言代码
│   ├── cmd/
│   │   ├── main.go       # 应用主入口
│   │   └── uploads/      # 上传文件存储目录（dish、ingredient、product子目录）
│   ├── config/           # 配置文件目录
│   │   ├── config_example.yaml
│   │   └── schema.sql
│   ├── internal/         # 内部包（领域模块化架构）
│   │   ├── auth/         # 认证模块（handler + service + dto）
│   │   ├── common/       # 公共模块
│   │   │   ├── dto/      # 数据传输对象（请求/响应结构体）
│   │   │   └── utils/    # 工具函数
│   │   │       ├── img/  # 图片处理工具
│   │   │       ├── jwt/  # JWT令牌工具
│   │   │       └── tempfs/# 临时文件系统
│   │   ├── config/       # 配置解析
│   │   ├── dish/         # 菜品模块（handler + service）
│   │   ├── ingredient/   # 食材模块（handler + service）
│   │   ├── middleware/   # 中间件（JWT认证）
│   │   ├── model/        # 数据模型（GORM实体）
│   │   ├── product/      # 产品模块（handler + service）
│   │   └── qrcode/       # 二维码模块（handler）
│   ├── go.mod            # Go模块定义
│   └── go.sum            # 依赖版本锁定
├── frontend/             # 前端Vue3代码
│   ├── src/
│   │   ├── api/          # API请求封装（拦截器、守卫）
│   │   ├── assets/       # 静态资源
│   │   ├── components/   # 通用组件（卡片、图片、Excel、二维码等）
│   │   ├── stores/       # Pinia状态管理
│   │   ├── types/        # TypeScript类型定义
│   │   └── views/        # 页面视图（admin管理端 + user用户端）
│   ├── index.html        # 入口HTML
│   └── package.json      # 前端项目配置
├── docker-compose.yml    # Docker编排配置
└── README.md             # 项目说明文档
```

## 技术栈

### 后端技术
- **编程语言**: Go 1.27
- **Web框架**: Gin 1.12.0
- **ORM框架**: GORM 1.31.1
- **数据库**: PostgreSQL
- **缓存**: Redis
- **认证**: JWT (JSON Web Token)
- **配置管理**: Viper 1.21.0
- **Excel处理**: Excelize 2.10.1
- **二维码生成**: go-qrcode
- **UUID生成**: google/uuid
- **图片处理**: disintegration/imaging
- **Web服务封装**: go-webservice

### 前端技术
- **框架**: Vue 3
- **语言**: TypeScript
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **网络请求**: Axios
- **拖拽功能**: vuedraggable

## 核心功能

### 1. 菜品管理 (Dish)
- 创建、编辑、删除菜品
- 上传菜品图片
- 管理菜品的食材组成
- 导出菜品数据到Excel

### 2. 食材管理 (Ingredient)
- 创建、编辑、删除食材
- 上传食材图片
- 从Excel导入食材数据
- 导出食材数据到Excel

### 3. 产品管理 (Product)
- 创建、编辑、删除产品
- 上传产品图片
- 关联食材信息
- 从Excel导入产品数据
- 导出产品数据到Excel

### 4. 认证系统 (Auth)
- 基于JWT的身份验证
- 权限控制

### 5. 二维码生成 (QRCode)
- 生成二维码图片

### 6. 图片处理
- 图片上传
- 图片压缩和调整尺寸

### 7. 缓存系统
- Redis缓存支持
- 数据库查询缓存优化

## 数据库结构

### 主要表结构

#### products 表
- 存储产品信息，包含产品编码、名称、价格、过敏原类型等

#### ingredients 表
- 存储食材基本信息

#### dishes 表
- 存储菜品信息，包含菜品编码、名称、做法等

#### dish_ingredients 表
- 存储菜品与食材的关联关系，包含用量和烹饪说明

#### 图片表
- dish_images: 菜品图片
- ingredient_images: 食材图片
- product_images: 产品图片

## 快速开始

### 环境要求
- Go 1.27+
- PostgreSQL
- Redis
- Node.js 18+

### 后端部署

1. **配置数据库**
   - 确保 PostgreSQL 服务已启动
   - 参考 `backend/config/config_example.yaml` 创建 `backend/config/config.yaml`
   - 修改数据库连接信息（host、port、user、password、dbname）

2. **初始化数据库**
   - 运行应用时会自动执行 `backend/config/schema.sql` 初始化数据库结构

3. **启动后端服务**
   ```bash
   cd backend
   go run cmd/main.go
   ```

### 前端部署

1. **安装依赖**
   ```bash
   cd frontend
   npm install
   ```

2. **启动开发服务器**
   ```bash
   npm run dev
   ```

3. **构建生产版本**
   ```bash
   npm run build
   ```

## API 接口

### 认证接口
- `POST /api/auth/admin/login` - 管理员登录

### 菜品接口
- `GET /api/dishes` - 获取菜品列表（游标分页）
- `POST /api/dishes` - 创建菜品（需认证）
- `GET /api/dishes/:dishCode` - 获取菜品详情
- `PATCH /api/dishes/:dishCode` - 更新菜品（需认证）
- `DELETE /api/dishes/:dishCode` - 删除菜品（需认证）
- `GET /api/dishes/:dishCode/ingredients` - 获取菜品关联食材列表
- `GET /api/dishes/export` - 导出菜品到Excel（需认证）

### 食材接口
- `GET /api/ingredients` - 获取食材列表（游标分页）
- `POST /api/ingredients` - 创建食材（需认证）
- `GET /api/ingredients/:ingredientCode` - 获取食材详情
- `PATCH /api/ingredients/:ingredientCode` - 更新食材（需认证）
- `DELETE /api/ingredients/:ingredientCode` - 删除食材（需认证）
- `GET /api/ingredients/:ingredientCode/products` - 获取食材关联产品列表
- `POST /api/ingredients/import` - 从Excel导入食材数据（需认证）
- `GET /api/ingredients/export` - 导出食材到Excel（需认证）

### 产品接口
- `POST /api/products` - 创建产品（需认证）
- `POST /api/products/import` - 从Excel导入产品数据（需认证）
- `GET /api/products/export` - 导出产品到Excel（需认证）
- `GET /api/products/:productCode` - 获取产品详情
- `GET /api/products/:productCode/dishes` - 获取产品关联菜品列表
- `PATCH /api/products/:productCode` - 更新产品（需认证）
- `DELETE /api/products/:productCode` - 删除产品（需认证）

### 二维码接口
- `POST /api/qrcode?url=<目标URL>` - 生成二维码图片

## 项目特点

1. **领域模块化架构**
   - 按业务领域划分模块（dish、ingredient、product、auth、qrcode），每个模块独立包含 handler 和 service
   - 公共能力抽取到 common 包（dto、utils），减少代码耦合
   - 使用 go-webservice 封装库加速开发

2. **完整的CRUD操作**
   - 支持菜品、食材、产品的完整增删改查
   - 提供Excel导入/导出功能，方便数据管理
   - 支持游标分页查询，适合大数据量场景

3. **图片处理系统**
   - 集成图片上传和自动处理
   - 支持图片压缩和尺寸调整

4. **安全认证**
   - 基于JWT的身份验证
   - 权限控制保护API接口

5. **缓存优化**
   - Redis缓存支持
   - 数据库查询缓存，提升性能

6. **配置管理**
   - 使用Viper管理配置
   - 支持环境变量和配置文件

7. **数据库优化**
   - 合理的索引设计
   - 完整的数据库模式定义

8. **现代化前端**
   - 基于Vue 3和TypeScript
   - 响应式设计，良好的用户体验

9. **扩展性**
   - 模块化设计，易于扩展新功能
   - 清晰的代码结构和命名规范

## 技术特性

- **RESTful API设计** - 遵循RESTful规范的API接口
- **事务管理** - 确保数据操作的一致性
- **错误处理** - 统一的错误处理机制
- **日志记录** - 详细的操作日志
- **数据验证** - 输入数据的有效性验证
- **文件上传** - 支持多文件上传
- **图片处理** - 自动图片优化
- **二维码生成** - 集成二维码生成功能
- **缓存机制** - Redis缓存 + SingleFlight 防缓存击穿
- **熔断保护** - Redis熔断器，防止缓存雪崩
- **临时文件系统** - 上传文件暂存与自动清理

## 开发指南

### 后端开发

1. **添加新功能模块**
   - 在 `internal/` 下创建新模块目录（如 `internal/newfeature/`）
   - 在模块目录中创建 `handler.go`（路由注册 + 请求处理）和 `service.go`（业务逻辑）
   - 在 `internal/common/dto/` 中定义通用的请求/响应结构体
   - 在 `internal/model/` 中定义数据模型
   - 在 `cmd/main.go` 中初始化并注册路由

2. **数据库迁移**
   - 修改 `config/schema.sql` 文件
   - 重新启动应用自动执行数据库初始化

### 前端开发

1. **添加新页面**
   - 在 `src/components/` 中添加新组件
   - 在 `src/views/` 中按 admin/user 分类创建页面组件
   - 在 `src/router.ts` 中添加路由配置
   - 在 `src/stores/` 中添加状态管理
   - 在 `src/types/types.ts` 中定义 TypeScript 类型

2. **API调用**
   - 在 `src/api/` 中封装 API 请求
   - 使用拦截器统一处理认证和错误

## 许可证

MIT License

## 备注
- 项目中使用了封装库 `github.com/LouYuanbo1/go-webservice` 来加快开发速度。
链接：[go-webservice](https://github.com/LouYuanbo1/go-webservice)。
欢迎使用这个封装库，并为其添加star，支持作者的创作，并提出宝贵的建议和问题。