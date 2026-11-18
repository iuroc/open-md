# OpenMD 项目约定

## 每次完成修改后

1. 先在当前分支提交所有已完成的改动，写清提交信息。除非明确要求，不要 push，也不要改写到已有提交。
2. 提交后再执行 `build.bat`，确保测试、`OpenMD.exe` 和 `source.git.bundle` 都是最新的。先提交再 build，bundle 才会包含最新提交。
3. `build.bat` 末尾有 `pause`，自动化调用时使用非交互方式：

   ```bash
   printf '\n' | cmd.exe //c build.bat
   ```

4. `OpenMD.exe` 和 `source.git.bundle` 已被 `.gitignore` 忽略，不要提交构建产物。
5. 如果测试或构建失败，先修复、重新提交，再报告完成；不要带着失败结果收尾。
6. 没有完成提交和构建时，不要把工作描述为已完成。
