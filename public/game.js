
//Animation and game updates will be triggered once connection made in index.hmtl

var animate = window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || function (callback) {
        window.setTimeout(callback, 1000/60)
    };
var canvas = document.createElement("canvas");
var width = 500;
var height = 500;
canvas.width = width;
canvas.height = height;
var context = canvas.getContext('2d');




var keysDown = {};     //This creates var for keysDown event

var playerScores = [];  //Each Player will use their playerTag from the server
                        //as the index of where to store their score
                        //The server will send the scores array over TCPChan

var playerX = 0;
var playerY = 0;
var playerHeight = 25;
var playerWidth = 50;
var playerSpeedX = 0;

var ballX = 0;
var ballY = 0;
var ballSpeedX = 0;
var ballSpeedY = 3;

var interpolateCounter = 0;

var render = function () {

   //Draw Background
   context.fillStyle = "#FF00FF";
   context.fillRect(0, 0, width, height);

   //Draw Players
   if(interpolation != undefined && previousUpdates != undefined && Updates != undefined && interpolateCounter >= 0 && Object.keys(previousUpdates).length == Object.keys(Updates).length){

     for (var key in interpolation) {     //interpolation defined in index.html
        if (key == "ball")  {
          //Ball
          context.beginPath();
          context.arc(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter, previousUpdates[key][1] + interpolation[key][1]*interpolateCounter, 5, 2 * Math.PI, false);
          context.fillStyle = "#000000";
          context.fill();
        }else{
          context.fillStyle = "#0000FF";
          context.fillRect(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter, previousUpdates[key][1] + interpolation[key][1]*interpolateCounter, 50, 50);
          //console.log(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter)
        }
     }

     if(interpolateCounter < interpolateFrames ){
       interpolateCounter++;
     }else{
       interpolateCounter = -1;
     }

   }else{
     for (var key in Updates) {     //Updates defined in index.html
        if (Updates.hasOwnProperty(key))  {
          context.fillStyle = "#FF00FF";
          context.fillRect(Updates[key][0], Updates[key][1], 50, 50);
          //console.log("oh dear")
        }
     }
   }

   //Local Player
   context.fillStyle = "#F7FF0F";   //yellow
   context.fillRect(playerX, playerY, 50, 50);
   //context.drawImage(pew, playerX, playerY, 200, 100, 100,100,50,50);
  //context.drawImage(pew, playerX, playerY, 200, 100);



};




var update = function() {
keyPress();

//check for ball hitting this player's paddle (client prediction)
/*
if (Math.abs( playerX - ballX ) < playerWidth && Math.abs( playerY - playerWidth/2 - ballY ) < playerHeight/2) {
      ballSpeedY = 3;
      ballSpeedX += (playerSpeedX / 2);
      //this.y += this.y_speed;
}
*/

};



var step = function() {
update();
render();
animate(step);
};



var keyPress = function() {

for(var key in keysDown) {
    var value = Number(key);

   if (value == 37) {   //37 = left
     playerX = playerX - 4;
     playerSpeedX = - 4;
     TCPChan.send("X" + playerX);
     TCPChan.send("S" + playerSpeedX);
	    //Put stuff for keypress to activate here
   } else if(value == 39){  //38 = right
     playerX = playerX + 4;
     playerSpeedX = + 4;
     TCPChan.send("X" + playerX);
     TCPChan.send("S" + playerSpeedX);
   } else if(value == 40){  //40 = down
     playerY = playerY + 4;
     TCPChan.send("Y" + playerY);
   } else if(value == 38){  //39 = up
     playerY = playerY - 4;
     TCPChan.send("Y" + playerY);
   }

 }

};






document.body.appendChild(canvas);



window.addEventListener("keydown", function (event) {
keysDown[event.keyCode] = true;
});


window.addEventListener("keyup", function (event) {
    delete keysDown[event.keyCode];
    if (event.keyCode == 37 || event.keyCode == 39){
      playerSpeedX = 0;
      TCPChan.send("S" + playerSpeedX);
    }

});
