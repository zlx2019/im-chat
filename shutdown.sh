# 要查询的进程名称
name="chat"

# 根据进程名 获取进程pid
pid=$(ps aux | grep "chat" | grep -v grep | awk '{print $2}')

if [ -z "$pid" ]; then
    # 进程未找到
    echo "进程未运行或者为空"
else
    # 关闭进程 或者直接根据进程名关闭进程 pkill [进程名]
    kill -9 $pid
    echo "shutdown process successfully"
fi
