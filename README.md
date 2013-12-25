Starry
======

A starbound server manager that adds ingame commands, item giving, banning and a couple of other features.

Source is on GitHub and has the most upto date information: https://github.com/d4l3k/starry

Features
-----
* Interactive CLI
* Ban players by IP or IP range.
* Player Join/Leave messages (see pictures below).
* Send messages from console.
* View logs
* Server restart on crash
* Works on Linux and Windows (not tested)
* `/command` commands. The syntax is the same as the CLI but you need to be on the admin list. Eg. `/ban Password1 Billy`. See `/help` for more information.
* Give players items.
* Admin chat highlighting.
* Server MOTD
* UUID Admins
* Loaded world monitoring `worlds`

Pretty Pictures
------
Join/Leave messages, MOTD and admin chat highlighting:

![Join Messages](http://i.imgur.com/77nJAAI.png)

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
<General>
  clients 
    - Connected client information (UUID, IP).
  players 
    - Show online players
  say <sender name> <message>
    - Say something.
  broadcast <message>
    - Show grey text in chat.
  color <color> <message>
    - Similar to broadcast but with color.
  help [<command>]
    - Information on commands.
  log [<count>]
    - Last <count> or 20 log messages.
  nick <name>
    - Change your character's name. In game only.
  item <name> <item> <count>
    - Give items to a player
  motd 
    - View the MOTD
  setmotd <message>
    - Sets the MOTD
  worlds 
    - List loaded worlds.
<Bans>
  bans 
    - Show ban list.
  ban <name>
    - Ban an IP by player name.
  banip <ip> [<name>]
    - Ban an IP or range (eg. 8.8.8.).
  unban <name>
    - Unban an IP by name.
  unbanip <ip>
    - Unban an IP.
<Admin>
  admins 
    - Lists the admins.
  adminadd <name>
    - Adds a player to the admin list.
  adminrem <name>
    - Removes a player from the admin list.
```

Future
-----
* Remote control via web.

Needed Software
-----
Starry is written in Go(lang) which is availible in most Linux distributions. Use your package manager to install it.

Links to downloads are provided here for Windows, Mac OSX and other Linux distributions:

http://golang.org/doc/install

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
Password: A password for remote admin access. Not used ATM. Will be used for the web interface.
MOTD: A message to display to users upon connection. Leave blank to disable.
Admins: A list of admins. Can be added to using 'addadmin' and 'deladmin'.
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
