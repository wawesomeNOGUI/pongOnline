
//Animation and game updates will be triggered once connection made in index.hmtl

var animate = window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || function (callback) {
        window.setTimeout(callback, 1000/60)
    };
var canvas = document.createElement("canvas");
var width = 400;
var height = 600;
canvas.width = width;
canvas.height = height;
var context = canvas.getContext('2d');


var fontReady = false;

//Get font
var f = new FontFace("Press Start 2P", 'url(https://fonts.gstatic.com/s/pressstart2p/v9/e3t4euO8T-267oIAQAu6jDQyK3nYivN04w.woff2) format(\'woff2\')');
//var f = new FontFace("Press Start 2P", 'url(https://fonts.googleapis.com/css?family=Press+Start+2P&display=swap) format(\'woff2\')');

  f.load().then(function(e) {
    // Ready to use the font in a canvas context
    console.log(e);
    document.fonts.add(e);
    fontReady = true;
});




var keysDown = {};     //This creates var for keysDown event

var playerScores = [];  //Each Player will use their playerTag from the server
                        //as the index of where to store their score
                        //The server will send the scores array over TCPChan

var playerX = 0;
var playerY = 0;
var pHeight = 25;
var pWidth = 50;
var playerSpeedX = 0;
var playerSpeedY = 0;

var ballX = 0;
var ballY = 0;
var ballSpeedX = 0;
var ballSpeedY = 3;

var interpolateCounter = 0;

var render = function () {

   //Draw Stuff From Server Updates Object
   if(interpolation != undefined && previousUpdates != undefined && Updates != undefined && interpolateCounter >= 0 && Object.keys(previousUpdates).length == Object.keys(Updates).length){
     //Draw Background
     context.fillStyle = "#FF00FF";
     context.fillRect(0, 0, width, height);

     //Draw Net
     context.fillStyle = "#FFFF00";
     context.fillRect(0,300,width,5);

     // Draw Players
     for (var key in interpolation) {     //interpolation defined in index.html
        if (key == "ball")  {
          //Ball
          context.beginPath();
          context.arc(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter, previousUpdates[key][1] + interpolation[key][1]*interpolateCounter, 5, 2 * Math.PI, false);
          context.fillStyle = "#000000";
          context.fill();
        }else if (key != playerTag){ //(don't redraw yourself) playerTag defined in index
          context.fillStyle = "#0000FF";
          context.fillRect(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter, previousUpdates[key][1] + interpolation[key][1]*interpolateCounter, pWidth, pHeight);
          //console.log(previousUpdates[key][0] + interpolation[key][0]*interpolateCounter)
        }
     }

     if(interpolateCounter < interpolateFrames ){
       interpolateCounter++;
     }else{
       interpolateCounter = -1;
     }

     //Draw Scores
     if(fontReady){
       context.font = "32px monospace Press Start 2P";  //We import PressStart2P in style.css
       context.fillStyle = "#000000";
       context.fillText(Updates["scoreT"], 10, 275);
       context.fillText(Updates["scoreB"], 10, 350);
     }


   }else{
     //draw nothing
   }

   //Local Player
   context.fillStyle = "#F7FF0F";   //yellow
   context.fillRect(playerX, playerY, pWidth, pHeight);
   //context.drawImage(pew, playerX, playerY, 200, 100, 100,100,50,50);
  //context.drawImage(pew, playerX, playerY, 200, 100);



};




var update = function() {
keyPress();
/*
//check for ball hitting this player's paddle (client prediction)
if (Updates != undefined){
  if (Math.abs( playerX - Updates.ball[0] ) < pWidth && Math.abs( playerY - pWidth/2 - Updates.ball[1]) < pHeight/2) {
    //Move ball (drawing only, doesn't affect real ball position server side)
        Updates.ball = [playerX, playerY];
  }
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
     playerSpeedX = -4;
     TCPChan.send("X" + playerX);
     TCPChan.send("SX" + playerSpeedX);
   } else if(value == 39){  //39 = right
     playerX = playerX + 4;
     playerSpeedX = 4;
     TCPChan.send("X" + playerX);
     TCPChan.send("SX" + playerSpeedX);
   } else if(value == 40){  //40 = down
     playerY = playerY + 4;
     playerSpeedY = 4;
     TCPChan.send("Y" + playerY);
     TCPChan.send("SY" + playerSpeedY);
   } else if(value == 38){  //38 = up
     playerY = playerY - 4;
     playerSpeedY = -4;
     TCPChan.send("Y" + playerY);
     TCPChan.send("SY" + playerSpeedY);
   }

 }

};






document.body.appendChild(canvas);



window.addEventListener("keydown", function (event) {
keysDown[event.keyCode] = true;
});


window.addEventListener("keyup", function (event) {
    delete keysDown[event.keyCode];
    if (event.keyCode == 37 || event.keyCode == 39) {
      playerSpeedX = 0;
      TCPChan.send("SX" + playerSpeedX);
    }
    if (event.keyCode == 38 || event.keyCode == 40) {
      playerSpeedY = 0;
      TCPChan.send("SY" + playerSpeedY);
    }

});
