# beehlp backend api server

## Starting and stopping servers
</br>

### Local

#### Start server

|          |                             |
|----------|-----------------------------| 
| Redis    | `brew services start redis` |
| backend  | `go run .`                  |
| frontend | `go run .`                  |


#### Stop servers
|          |                             |
|----------|-----------------------------| 
| frontend | `Ctrl-c`                    |
| backend  | `Ctrl-c`                    |
| Redis    |`brew services stop redis`   |

</br>

### Google CLoud Platform

see ~/.bashrc for aliases

</br>

#### Start server
|          |                                            |
|----------|--------------------------------------------| 
| Redis    | `sudo service redis-server start`          |
| backend  | `sudo rm nohup.out;nohup sudo -b go run .` |
| frontend | `sudo rm nohup.out;nohup sudo -b go run .` |

</br>

#### Stop servers
|          |                                  |
|----------|----------------------------------| 
| frontend | `sudo killall bhfeprototype`     |
| backend  | `sudo killall bhbe`              |
| Redis    | `sudo service redis-server stop` |
| log      | `tail -f nohup.out`              |

