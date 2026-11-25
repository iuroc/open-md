# Spec Delta

## REMOVED Requirements

### Requirement: Outline navigation leaves source mode

**Reason**: 源码模式下点击大纲不应强制切回阅读模式；用户需要在源码中直接定位到对应标题行。新行为由 `outline-navigation` 能力定义。

**Migration**: 无需迁移。点击大纲将保持在源码模式并滚动到对应标题行；用户仍可手动切回阅读模式。
