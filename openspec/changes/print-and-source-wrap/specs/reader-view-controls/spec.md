# Spec Delta

## Purpose

定义阅读页顶栏阅读模式与宽度控件在不同阅读模式下的可用性和可见性，避免显示对当前模式无效的控件。

## ADDED Requirements

### Requirement: Split view hides the width toggle

双视图模式下，顶栏的宽度模式按钮 MUST NOT 显示。切回阅读模式或源码模式时，该按钮 SHALL 恢复显示。用户在阅读模式中选择的全宽偏好 MUST 保留，不因进入双视图而重置。

#### Scenario: Width toggle hidden in split view

- **WHEN** 用户从阅读模式切换到双视图
- **THEN** 顶栏的宽度模式按钮不再显示

#### Scenario: Width toggle restored after leaving split view

- **WHEN** 用户从双视图切回阅读模式或源码模式
- **THEN** 宽度模式按钮恢复显示

#### Scenario: Full-width preference is retained

- **WHEN** 用户在阅读模式启用全宽后进入双视图，再切回阅读模式
- **THEN** 阅读模式仍保持全宽状态

#### Scenario: Narrow viewport falls back to read mode

- **WHEN** 窗口宽度不足导致双视图回退为阅读模式
- **THEN** 宽度模式按钮正常显示
