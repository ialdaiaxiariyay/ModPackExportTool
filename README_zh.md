### 中文文档

pack-export 是一个用 Go 编写的命令行工具，用于将 Minecraft 游戏实例导出为 Modrinth 兼容的整合包格式 (.mrpack)。它会扫描游戏目录，根据可配置规则选择性收集文件，查询 Modrinth 和 CurseForge 的 API 来将本地文件映射到远程下载地址，并生成 `modrinth.index.json` 清单文件。未匹配或未查询的文件会直接复制到 `overrides` 文件夹中。该工具还支持可选的 Git 仓库初始化和最终打包为 `.mrpack` 压缩包。

#### 功能特性

- 可配置的文件选择，涵盖游戏核心、设置、模组、配置、资源包、光影、存档以及许多模组专用数据文件夹。
- 从版本 JSON 中自动检测 Minecraft 版本、Forge、Fabric 和 NeoForge。
- 通过网络查询 Modrinth (SHA1) 和 CurseForge (MurmurHash2 指纹) 获取远程下载链接。
- 支持跳过网络模式（离线或直接打包，注意：可能违反二次分发协议）。
- 在输出目录中可选初始化 Git 仓库，支持设置远程地址和分支。
- 支持向压缩包根目录复制额外文件或文件夹。
- 生成带完整注释的配置模板。

#### 使用方法

```bash
pack-export -config <配置文件.yaml> [选项]
```

| 参数 | 说明 |
|------|------|
| `-config` | YAML 配置文件路径（除非使用 `-save-config`，否则为必需）。 |
| `-output` | 输出目录路径（默认为 `名称_版本`）。 |
| `-name` | 覆盖配置文件中的整合包名称。 |
| `-version` | 覆盖配置文件中的整合包版本。 |
| `-init-git` | 在输出目录中初始化 Git 仓库（覆盖配置中的 `git.init`）。 |
| `-package` | 将导出的目录打包为同级的 `.mrpack` 文件。 |
| `-save-config` | 生成带注释的 YAML 配置模板并退出。 |

示例：

```bash
pack-export -config mypack.yaml -output MyPack_1.0 -package
```

#### 配置说明

YAML 配置文件控制导出行为。可使用 `-save-config` 生成完整的模板。主要部分包括：

- `game_path`：Minecraft 实例目录的绝对或相对路径（必须是隔离的版本文件夹或 `.minecraft` 根目录）。
- `export`：整合包名称、版本以及每个文件类别的精细布尔选项（如 `basic`、`mod`、`config`、`resource_packs`、`saves` 等）。
- `advanced`：`skip_network`（直接复制所有文件）和 `modrinth_only`（跳过 CurseForge 查询）。
- `pcl`：PCL 可执行文件路径（如果包含启动器）。
- `git`：`init`、`remote` 和 `branch` 设置。
- `rules_overrides` / `extra_files`：高级包含/排除规则以及要放置在压缩包根目录的额外文件。

#### 构建

```bash
go build -o pack-export main.go
```

#### 参考

本项目参考了 [Meloong-Git/PCL](https://github.com/Meloong-Git/PCL) 的部分实现。