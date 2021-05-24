# PowerMock

PowerMock是一个Mock Server的实现，它同时支持HTTP与gRPC协议接口的Mock，并提供了灵活的插件功能。

## 功能

作为一个Mock Server，PowerMock具有以下的核心功能：
1. 支持HTTP协议与gRPC协议接口的Mock。
2. 支持配置Javascript等脚本语言来动态生成响应。
3. 支持对一个接口配置多种响应，并按照条件进行区分。
4. 匹配条件支持多种运算符(AND/OR/>/</=等)。
4. 支持返回静态数据以及特定领域的随机数据。
5. 支持插件功能，可以通过编写插件实现其他匹配或Mock引擎。
6. 同时提供HTTP与gRPC接口，可以动态对MockAPI进行增删改查。

## 示例

