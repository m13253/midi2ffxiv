MIDI2FFXIV
==========

《最终幻想 14：红莲之狂潮》版本吟游诗人 MIDI 演奏工具，可多人同步演奏。

使用方法
-----

下载地址：[midi2ffxiv-########.zip](https://github.com/m13253/midi2ffxiv/releases)。

当前版本支持 64 位 Windows 系统下的 FFXIV 4.3 版本。

启动后按照以下步骤开启控制面板，您将会看到：

![Screenshot](screenshot.png)

您也可以通过本机的 IP 地址在另一台设备上开启控制面板。

演奏模式
-----

- 手动独奏
- MIDI 自动演奏
- 多人合奏

（请注意：多人合奏模式下，演奏时间会与路人同步以达到更好的演出效果，演奏者自己的演出效果则有可能不佳。）

特性
--------

- 网页客户端：可以用手机或另一台电脑远程控制（由于内存消耗等因素，在某些情况下并不推荐）
- 125 毫秒延迟队列：自动分解和弦
- 本地 MIDI 回放：在合成器延迟低的情况下可以试听演奏效果
- 修复了 4.3 版本中某些音符不能弹奏的问题
- NTP 时钟同步：和亲朋好友在游戏中合奏！
- 手动时钟同步：好的乐队指挥是成功的一半
- 定时自动演奏/循环演奏
- 多人合奏 1500ms 延迟（不了解的话可以在实际操作中明白）

视频展示
--------

"Saltswept" 二重奏: <https://www.youtube.com/watch?v=2n6HCc1FdsQ>

"Prelude" 二重奏: <https://www.youtube.com/watch?v=yINX3F7jKkU>

键位设置
----------

您有两种键位设置可以选择，一种是默认的 `midi2ffxiv.conf`，还有另外的 `midi2ffxiv_no_modifier.conf`。

- `midi2ffxiv.conf` 使用 Ctrl 和 Shift 切换八度，在某些特定情况下可能导致帧率降低。
- `midi2ffxiv_no_modifier.conf` 使用全键盘键位。请按照以下表格设置：

|      | C | D  | E  | F  | G  | A | B | C+1 |
|------|---|----|----|----|----|---|---|-----|
|  高  | A | D  | G  | H  | K  | ; | - | =   |
|  中  | Q | W  | E  | R  | T  | Y | U |     |
|  低  | Z | C  | B  | N  | ,  | / | ] |     |
|      |   |    |    |    |    |   |   |     |
|      | C | Eb | F# | G# | Bb |   |   |     |
|  高  | S | F  | J  | L  | '  |   |   |     |
|  中  | 2 | 3  | 5  | 6  | 7  |   |   |     |
|  低  | X | V  | M  | .  | \[ |   |   |     |

无论使用哪种配置，将其改名为 `midi2ffxiv.conf` 后才能生效。

手动独奏模式
----------------

如果您想用 MIDI 键盘或 MIDI 控制器演奏，请在“输入设备”里选择你的 MIDI 设备。

另外：如果您想用本地回放功能（见下文），请在“输出设备”里选择你的合成器、乐器，并将您的 MIDI 合成器音量调节到一个适宜的程度，以便您能同时听清游戏和合成器的声音。

现在可以开始您的表演了！请不要过快弹奏，如果音符间的时间间隔超过 125ms，可能会出现延迟或丢失个别音符。

（注意：MIDI2FFXIV 将实时演奏的音符间距限制在了 125ms 以上，游戏中也是相同的配置，不过您可以在 [midi2ffxiv.conf](midi2ffxiv.conf) 中更改）。

MIDI 自动演奏模式
------------------

首先请加载 MIDI 文件，[demo](demo) 文件夹下有一些示例文件。然后选择音轨。

如果您选择了音轨 1，播放的却是音轨 2，请尝试使用隐藏值“Track 0”。

选择音轨后请点击“当前时间”旁的“复制”，然后点击“开始时间”旁的“设置”。5 秒钟后演奏即会开始。

再点一次“设置”即可终止（可用另一台电脑操作）。您也可以用“Ctrl-Alt-Shift-\[”强行中止。

（注意：部分网上下载的 MIDI 文件可能无法在 MIDI2FFXIV 中播放。如果您有编曲软件，可以考虑自制 MIDI 文件。）

多人合奏模式
---------------------

首先，点击“NTP 服务器”旁的“同步”，等待 5 到 10 秒后同步成功。

加载**预演 MIDI 文件**，选择音轨。

与指挥设定好时间后，将时间填入“起始时间”，点“设置”启动定时器。

预演过程中指挥可以调节每个人的“偏移”，让每个声部同步演奏。

再次点击“设置”停止播放，然后请加载**正式演出 MIDI 文件**。

依照正式演出的起始时间设置定时器。

（注 1：4.3 版本后，演奏者与观众的延迟大约有 1500ms，MIDI2FFXIV 在非实时演奏模式中模拟了这一延迟。）

（注 2：指挥很重要！至少需要三个人（一个指挥和两个以上演奏者）才能调整各演奏者之间的同步设置。）

本地回放
----------

该功能为次要功能。如果你的合成器延迟比游戏低，你可以用这个试听。

通常用在你有硬件合成器的时候，当然，软件合成器也可以用。（比如 [VirtualMIDISynth](https://coolsoft.altervista.org/en/virtualmidisynth)）

如果用了 VirtualMIDISynth，你可以将缓冲时间设定为 5 到 10ms，从而降低延迟。

常见问题
---

1. **如何更改键位设置？**

   默认键位在 [midi2ffxiv.conf](midi2ffxiv.conf) 里设置，可用记事本更改。

   注意：对于非数字键/字母键，请根据此表格查询键盘码：[Virtual-Key Codes](https://docs.microsoft.com/en-us/windows/desktop/inputdev/virtual-key-codes)。

2. **使用 MIDI2FFXIV 会被封号吗？**

   应该不会。目前还没有见过任何禁用 MIDI 的公告和声明。

   不过请记住，不要加载各种奇葩 MIDI 文件，这会给服务器造成负担。也不要上传《Answers》《Dragonsong》《Revolutions》的视频，这些会有侵权问题。

   还有请在您的视频中加入以下版权声明：
   ```
   FINAL FANTASY XIV © 2010 - 2018 SQUARE ENIX CO., LTD. All Rights Reserved.
   ```

3. **为什么 MIDI2FFXIV 需要管理员权限？**

   删除 `midi2ffxiv.exe.manifest` 即可。

   不过，如果您的客户端是在 UAC 环境下运行的（如 FFXIV 国服），那么 MIDI2FFXIV 也需要管理员权限。

4. **MIDI2FFXIV 需要防火墙放行吗？**

   如果您需要用手机或另一台电脑控制 MIDI2FFXIV，则需要防火墙放行。如果你的电脑直连外网，建议加上密码。

   在 [midi2ffxiv.conf](midi2ffxiv.conf) 中找到以下内容并更改：

   ```conf
   WebUsername
   WebPassword
   ```

   在这里加上你的用户名和密码。

5. **我的杀毒软件对 MIDI2FFXIV 报毒！**

   我的也有这种情况。

   如果您不信任已编译的程序，您可以自行编译（见下）。

6. **如何自行编译 MIDI2FFXIV？**

   你需要下载 [Go](https://golang.org/dl/)。

   在命令提示符下输入以下命令编译：

   ```cmd
   cd /d "SOURCE CODE PATH"
   go get -d -u -v .
   go build
   ```

版权声明
-------

This program is licensed under MIT License.

For more information, please refer to [LICENSE](LICENSE).

FINAL FANTASY is a registered trademark of Square Enix Holdings Co., Ltd.

Demo songs in [demo](demo) directory may have separate licensing information, please refer to [demo/README.txt](demo/README.txt).
