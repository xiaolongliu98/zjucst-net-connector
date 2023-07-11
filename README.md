# zjucst-net-connector
快速连接校园网的工具

# 环境变量配置
使用前请在环境变量中填写配置 

CST_USERNAME_CONFIG：[WIFI_NAME1]:[Username1]【,[Password1]】;[WIFI_NAME2]:[Username2]【,[Password2]】;

其中【】是可选的，如果您的密码不是默认密码zjucst则需要填写Password，Password是通过某种加密算法进行加密的，因此需要通过抓取您实际登录时login的HTTP请求中进行获取
