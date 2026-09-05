#!/bin/sh
cd /app || exit

# 自动更新支持：如果存在新版本二进制，替换当前版本
if [ -f "./dockpit-new" ]; then
    mv ./dockpit-new ./dockpit
    chmod +x ./dockpit
fi

# 运行 dockpit
exec ./dockpit
