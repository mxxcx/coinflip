### Coinflip Game Backend API ###

 **GET /api/coinflipws ** 

Web Sockets endpoint, gets push notifications on new games, completed games, deleted games

 **GET /api/coinflip ** 
   
List all the open games, no jwt required.
   
   **POST /api/coinflip ** 
   
jwt required, creates new game, takes json in the request body, eg. {"pick": "red", "amount": 10} pick can be red or blue.

**PUT  /api/coinflip/:gameId** 
   
jwt required, joins a game, based on the gameID eg. PUT /api/coinflip/26 joins the game with the ID 26.

**DELETE  /api/coinflip/:gameId** 
   
jwt required, deletes a game, based on the gameID eg. DELETE /api/coinflip/26 deletes the game with the ID 26.

 **GET /api/coinflip-history ** 
   
jwt required, gets the jwt's user games both created or joined. 

 **GET /api/coinflip-top-players ** 
   
jwt not required, gets the top 2 player by amount won 
