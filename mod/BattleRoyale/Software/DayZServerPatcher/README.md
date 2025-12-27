# DayZServerPatcher

Patches DayZ server to enable write operations to the root directory.

## Developer Notes

In order to manage global bans for BattleRoyale, we needed to write to the
root bans.txt file. DayZ server does not allow this by default, so this patch
enables that functionality.

You can read more about how this patch works [on my blog](https://blog.lystic.dev/2020/10/24/enabling-file-write-to-dayz-root-directory/).
