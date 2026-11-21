# OpenMD 安装弹窗（TaskDialog）设计

日期：2026-11-21
状态：已确认

## 目的

把双击 `OpenMD.exe`（无参数启动）时的 MessageBox 是/否/取消，换成 Windows 原生 TaskDialog。安装成功不再追加一个完成提示，整个启动流程最多只出现一个弹窗。

## 目标

- 未安装：双击后直接安装，安装成功后弹一个“OpenMD 已安装”对话框，提供“卸载 / 完成”。
- 已安装：双击后弹一个“OpenMD 已安装”对话框，提供“覆盖安装 / 卸载 / 取消”。
- 成功路径不再有第二个“安装完成”提示。
- 使用系统自带的 TaskDialog，不引入 WebView2、Electron 或额外运行时。

## 非目标

- 不改阅读页的浏览器打开方式，不做页面内卸载。
- 不改临时目录策略：仍写入安装目录下的 `openmd-temp`，不可写时回退 `%TEMP%\OpenMD`。
- 不改 `ConfirmLoad`（大量文档时的“加载全部 / 快速加载 / 取消”），后续可以复用同一套 TaskDialog。
- 不区分“更新”和“管理”两种已安装状态；只要已安装，统一显示“覆盖安装 / 卸载 / 取消”。

## 方案比较

- **A WebView2 弹窗（未采用）**：视觉上限最高，但引入 WebView2 运行时依赖；一个一次性安装弹窗不值得。部分 LTSC / Server 机器还可能没有运行时，需要双套 UI 兜底。
- **B 自绘 Win32 窗口（未采用）**：没有依赖，但 DPI、字体、无障碍、按钮状态都要自己维护，成本高于收益。
- **C 原生 TaskDialog（采用）**：系统组件、零额外体积，支持大标题、图标和带说明的按钮，和 Windows 原生体验一致。

### 关键实现坑

`TASKDIALOGCONFIG` 在头文件里是 **pack(1)** 结构：64 位下 `cbSize` 为 **160**，字段之间没有自然对齐填充；`TASKDIALOG_BUTTON` 也是 packed，每项 **12 字节**。Go 结构体无法表达 packed 布局，必须用字节缓冲区按偏移手工序列化，否则 `TaskDialogIndirect` 会返回 `E_INVALIDARG (0x80070057)`。

## 行为

### 未安装

1. 双击 `OpenMD.exe`。
2. 直接把当前 exe 复制到 `%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe`，并注册右键菜单。
3. 弹出一个 TaskDialog：

```text
┌─ OpenMD ──────────────────────────────────┐
│ [图标] OpenMD 已安装                       │
│                                            │
│ 右键 Markdown 文件或文件夹，选择 OpenMD     │
│ 即可阅读。                                  │
│                                            │
│ › 卸载                                     │
│ › 完成                                     │
└────────────────────────────────────────────┘
```

默认按钮是“完成”。选“卸载”会立即卸载并结束；选“完成”直接退出。

### 已安装

双击后先弹窗，不直接动文件：

```text
┌─ OpenMD ──────────────────────────────────┐
│ [图标] OpenMD 已安装                       │
│                                            │
│ 右键 Markdown 文件或文件夹，选择 OpenMD     │
│ 即可阅读。                                  │
│                                            │
│ › 覆盖安装                                  │
│ › 卸载                                     │
│ › 取消                                     │
└────────────────────────────────────────────┘
```

默认按钮是“取消”。“覆盖安装”调用现有 `Installer.Install`；如果当前运行的正是安装目录里的 exe，`Install` 已识别同路径并只刷新注册表。

### 错误

任何步骤失败只显示一个错误 TaskDialog。成功路径不再有 `Info` 提示。

### 命令行

- `OpenMD.exe --install`：静默安装，成功不弹窗，失败弹错误框。
- `OpenMD.exe --uninstall`：静默卸载，成功不弹窗，失败弹错误框。

## 技术设计

### 资源与清单

- 新增 `assets/app.manifest`：
  - Common Controls v6 依赖（TaskDialogIndirect 只在 v6 导出）。
  - `asInvoker`。
  - Windows 10 supportedOS。
  - PerMonitorV2 DPI，顺带修复高 DPI 下旧 MessageBox 的模糊。
- 用已有 `rsrc` 重新生成资源：

```bash
rsrc -ico assets/openmd.ico -manifest assets/app.manifest -o rsrc_windows_amd64.syso
```

`go build` 自动链接 `.syso`，日常构建不需要安装 `rsrc`。

### UI 层

- `internal/ui/taskdialog_windows.go`：
  - packed `TASKDIALOGCONFIG` / `TASKDIALOG_BUTTON` 序列化。
  - `TaskDialogIndirect` 调用；不可用时回退 MessageBox。
  - 应用图标用 `ExtractIconEx(exe, 0)` 取第一个图标组，避免依赖资源 ID。
  - 错误框使用系统错误图标。
- `internal/ui/dialog.go`：平台无关的接口与枚举。
- `internal/ui/dialog_windows.go`：`WindowsDialog` 实现。
- `Dialog` 接口调整为：

```go
type InstallMode int

const (
    InstallAfterInstall InstallMode = iota // 刚安装完，卸载 / 完成
    InstallManage                          // 已安装，覆盖安装 / 卸载 / 取消
)

type Dialog interface {
    Installer(mode InstallMode) Choice
    ConfirmLoad(message string) LoadDecision
    Error(message string)
}
```

删除 `Info`。

### 应用层

`internal/app/run.go` 的 `runInstaller`：

- `installed := fileExists(Installer.InstalledExe())`。
- 未安装：`Install(currentExe)` → `Dialog.Installer(InstallAfterInstall)` → 选卸载则卸载，否则退出。
- 已安装：`Dialog.Installer(InstallManage)` → 覆盖安装 / 卸载 / 取消。
- 成功路径不调用 `Info`。
- `handleInternal` 的 `--install` / `--uninstall` 同样删除成功提示。

## 测试

- UI 纯逻辑：
  - `InstallAfterInstall` 按钮为“卸载 / 完成”，默认“完成”。
  - `InstallManage` 按钮为“覆盖安装 / 卸载 / 取消”，默认“取消”。
  - 按钮 ID 映射：100 → 安装，101 → 卸载，其他 → 取消。
- 应用流程：
  - 未安装时先安装，再询问 `InstallAfterInstall`。
  - 已安装时询问 `InstallManage`，覆盖安装 / 卸载分别调用对应操作。
  - 卸载需要 helper 时调用 `SpawnUninstallHelper`，成功路径无 `Info`。
- 构建后用资源探针确认 `RT_MANIFEST` 在 exe 内，并实际打开三个状态检查观感。

## 限制

- TaskDialog 是系统原生样式，不跟随阅读页的四套配色，也不跟随深色模式。
- 阅读页在默认浏览器里，没有回调通道，无法从页面内发起卸载；卸载入口是安装目录里的 `OpenMD.exe` 或 `--uninstall`。
- 已安装状态不比较版本号；“覆盖安装”永远用当前 exe 覆盖。
