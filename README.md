# mossh                                                                   
                                                                     
This is a simple [ssh](https://github.com/charmbracelet/wish) wrapper around [charmbracelet/mods](https://github.com/charmbracelet/mods) designed to allow you to acccess your mods instances from anywhere

## Examples
Take a look at the mods examples to see what you can do:

https://github.com/charmbracelet/mods/blob/main/examples.md

Once running and exposed through port 22 you can run like so:

```
ssh mods.your.domain -t "whats up doc?"
```

or to enter a chat app:
```
ssh mods.your.domain
```

## How To                                         
See `docker-compose.yml` for details. 

By default it will accept all incoming connections with a valid public key. In order to whitelist only specific public keys, create an allowlist file and set the `MOSSH_ALLOW_LIST` env variable. 


## Local Testing
1. Create a `.env` with the values you want from [charmbracelet/mods docs](https://github.com/charmbracelet/mods)
2. `go run .`
3. Open another terminal and run: `ssh localhost -p 23234`


## License                                                             
Released under the MIT License. See the  LICENSE  file for more  
information.