# 第 10 章：组织一篇文档

单个文件能写清楚一件事，多个文件才能组织一个项目。这一章讲目录结构、入口文件和 front matter。

## 你会学到

- 用 `README.md` 或 `index.md` 作为目录入口。
- 用相对链接连接多个文档。
- 处理 YAML / TOML front matter，以及移动文件时的注意事项。

## 基本写法

在目录模式下，OpenMD 会递归扫描 Markdown 文件，并优先打开目录里的 `README.md` 或 `index.md`。一个常见的结构是：

```text
project/
  README.md
  docs/
    install.md
    usage.md
  CHANGELOG.md
```

入口文件可以这样组织：

```markdown
# 项目说明

一句话介绍项目。

## 文档

- [安装](docs/install.md)
- [使用](docs/usage.md)
- [更新日志](CHANGELOG.md)
```

有些工具会在文件开头写 front matter：

```markdown
---
title: 项目说明
tags:
  - 文档
---

# 项目说明
```

OpenMD 在阅读时会自动去掉开头的 YAML（`---`）或 TOML（`+++`）块；源码视图仍然显示它。

## 在 OpenMD 里怎么看

目录模式会按真实层级展开文件树；链接到其他 `.md` 文件时，侧栏会出现对应页面，可以直接跳转。根目录的 `README.md` / `index.md` 会作为初始页面。

移动或重命名文件后，原来的相对链接会失效。OpenMD 遇到打不开的链接不会崩溃，但读者会看到跳转失败，所以调整目录后要顺手搜索并更新链接。

## 常见错误

- 在根目录放多份入口文件：`README.md` 和 `index.md` 同时存在时，优先级容易让人困惑。
- 使用绝对路径：换一台电脑或换一个目录就打不开。
- front matter 不在文件最开头：不会被识别，反而会显示在正文里。
- 文件名大小写不一致：Windows 上可能正常，换到其他系统就失效。
- 移动文件后忘记改链接：跨文档导航断开。

## 延伸阅读

- [GitHub Docs：关于 README](https://docs.github.com/zh/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes)
- [Markdown Guide：文档结构建议](https://www.markdownguide.org/basic-syntax/)

上一章：[分隔线、HTML 与安全边界](09-rules-and-html.md) · 下一章：[五个实战配方](11-recipes.md)
