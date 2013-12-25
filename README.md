Starry
======

A starbound server manager that adds ingame commands, banning and a couple of other features.

Features
-----
* Interactive CLI
* Ban players by IP or IP range.
* Player Join/Leave messages (see pictures below).
* Send messages from console.
* View logs
* Server restart on crash
* Works on Linux and Windows (not tested)
* `/command` commands. The syntax is mostly the same as the CLI but you need to put the server password as the first argument for most commands. Eg. `/ban Password1 Billy`. See `/help` for more information.
* Give players items.
* Admin chat highlighting.

Pretty Pictures
------
Join/Leave messages and admin chat highlighting:
![Join Messages](http://i.imgur.com/rlnzsoV.png)

Ingame commands:
![Ingame](http://i.imgur.com/xq3lZK6.png)

Giving Items:
![Item](http://i.imgur.com/mCAWxE8.png)

Interactice CLI (old pic):
![CLI](http://i.imgur.com/ZKP9OHM.png)

Commands
-----
```
Starry CLI - Welcome to Starry!
Starry is a command line Starbound and remote access administration tool.
> help
[Commands]
General:
  clients 
    Display connected clients.
  say <sender name> <message>
    Say something.
  broadcast <message>
    Show grey text in chat.
  color <color> <message>
    Similar to broadcast but with color.
  help [<command>]
    Information on commands.
  log [<count>]
    Last <count> or 20 log messages.
  nick <name>
    Change your character's name. In game only.
  item <name> <item> <count>
    Give items to a player
Bans:
  bans 
    Show ban list.
  ban <name>
    Ban an IP by player name.
  banip <ip> [<name>]
    Ban an IP or range (eg. 8.8.8.).
  unban <name>
    Unban an IP by name.
  unbanip <ip>
    Unban an IP.
Admin:
  admins 
    Lists the admins.
  addadmin <name>
    Adds a player to the admin list.
  deladmin <name>
    Removes a player from the admin list.
```

Future
-----
* Remote control via web.

Needed Software
-----
Starry is written in Go(lang) which is availible in most Linux distributions. Use your package manager to install it.

Use
------

Modify `starbound.config` in your linux64(/32) folder and change `gamePort` to be 21024, or whatever port you would like. You should then modify your firewall to block this port. Alternatively, if you're selectively port-forwarding ports, don't port forward 21024. 

On Linux you can do this by running:
```bash
sudo iptables -A INPUT -p tcp --destination-port 21024 -j DROP
```

Modify `starry.config` to your prefered values.
```
ServerPath: You should modify "gamePort" in the starbound.config file in the ServerPath folder to be 21024.h to the starbound_server executable.
LogFile: This is the path to the log file location. If you leave this blank it will append ".log" to the ServerPath.
ServerAddress: Address that the Starbound server can be connected to at. 
ProxyAddress: The address that Starry binds to. This should probably be left as is.
Password: A password for remote admin access.
Admins: A list of admins. Can be added to using 'addadmin' and 'deladmin'. The only thing "admin" status gives you is green text in chat.
Bans: Leave this as is unless you know what you are doing. This is used by Starry to save the bans.
```

To launch it:
```bash
go run starry.go
```


License
-----
The code is free to use and licensed under the MIT License.

Copyright (c) 2013 Tristan Rice
