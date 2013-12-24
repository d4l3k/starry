Starry
======

A starbound server manager with an API for remote administration.

Features
-----
* Interactive CLI
* Ban players by IP or IP range.
* Player Join/Leave messages (see pictures below).
* Send messages from console.
* View logs
* Server restart on crash
* Works on Linux and Windows (not tested)

Future
-----
* `/command` commands.
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
Bans: Leave this as is unless you know what you are doing. This is used by Starry to save the bans.
```

To launch it:
```bash
go run starry.go
```

Pretty Pictures
------
Join/Leave Message
![Join Messages](http://i.imgur.com/jePE5aH.png)

Interactice CLI
![CLI](http://i.imgur.com/ZKP9OHM.png)

License
-----
The code is free to use and licensed under the MIT License.

Copyright (c) 2013 Tristan Rice
