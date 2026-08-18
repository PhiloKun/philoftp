PhiloFTP 内网 FTP 服务器
========================

PhiloFTP 是一个轻量的内网 FTP 服务器，内置 Web 管理端，带系统托盘图标常驻后台。

快速开始
--------
1. 安装完成后勾选"立即启动"，或在开始菜单/桌面点击 PhiloFTP 快捷方式
2. 程序以**系统托盘图标**方式常驻后台（任务栏右侧区域，▲ 展开可见青色 FTP 图标）
3. **左键/右键单击托盘图标**可弹出菜单：
   - 状态：显示当前运行状态
   - 打开 Web 管理页：在浏览器打开管理界面
   - 启动服务 / 停止服务
   - 打开日志目录
   - 退出 PhiloFTP
4. 浏览器访问 http://本机IP:9090 进入 Web 管理端
5. 默认登录：admin / admin123（请尽快修改）

常用端口
--------
- Web 管理端口：9090
- FTP 控制端口：2121
- FTP 被动端口范围：21100-21110

数据与日志位置
--------------
- 配置/用户数据库：%USERPROFILE%\.philoftp\configs\
- 用户文件：%USERPROFILE%\.philoftp\data\（若未指定）
- 运行日志：%USERPROFILE%\.philoftp\logs\philoftp.log

命令行模式（可选）
------------------
若需要控制台窗口（显示日志、方便调试），可在快捷方式目标后加参数：
  philoftp.exe -headless
  或自定义端口：philoftp.exe -web-port 8080 -ftp-port 21

系统要求
--------
Windows 10 / 11（64 位）
