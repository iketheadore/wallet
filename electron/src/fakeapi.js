const proxy = require('express-http-proxy');
const express = require('express')
const app = express()

app.all('/*', function(req, res, next) {
  res.header("Access-Control-Allow-Origin", "*");
  res.header("Access-Control-Allow-Headers", "X-Requested-With");
  next();
});

//Real API proxy
app.use('/api/*', proxy('127.0.0.1:8080', {
  proxyReqPathResolver: function(req) {
  	let url = "/api/" + req.params[0];
    return url;
  }
}));

const path = "/fakeapi/";
//Serve the static js files
app.use(express.static('starter-kit/dist'))

//Fake API endpoints
app.get(path + 'explore/all', (req, res) => {
	successResponse(res, generateCats(100));
});

app.get(path + 'boxes/all', (req, res) => {
	successResponse(res, generateCats(57));
});

app.get(path + 'user/:userId/mine', (req, res) => {
	successResponse(res, generateCats(2));
});

app.get(path + 'forsale/all', (req, res) => {
	successResponse(res, generateCats(25));
});

app.get(path + 'forsire/all', (req, res) => {
	successResponse(res, generateCats(33));
});

app.all('*', function(req, res) {
  res.redirect("/");
});


//Random generator code

function randomGeneration()
{
	return Math.floor(Math.random()*5);
}
function randomName()
{
	var catNames = require('cat-names');

	return catNames.random();
}
function randomBreed()
{
	//List from http://cattime.com/cat-breeds
	const breeds = [
		"Abyssinian", 
		"American Bobtail", 
		"American Curl", 
		"American Shorthair", 
		"American Wirehair", 
		"Balinese", 
		"Bengal Cats", 
		"Birman", 
		"Bombay", 
		"British Shorthair", 
		"Burmese", 
		"Burmilla", 
		"Chartreux", 
		"Chinese Li Hua", 
		"Colorpoint Shorthair", 
		"Cornish Rex", 
		"Cymric",
		"Devon Rex",
		"Egyptian Mau", 
		"European Burmese",
		"Exotic",
		"Havana Brown", 
		"Himalayan",
		"Japanese Bobtail",
		"Javanese",
		"Korat",
		"LaPerm",
		"Maine Coon",
		"Manx",
		"Nebelung",
		"Norwegian Forest",
		"Ocicat",
		"Oriental",
		"Persian",
		"Pixie-Bob",
		"Ragamuffin",
		"Ragdoll Cats",
		"Russian Blue",
		"Savannah",
		"Scottish",
		"Fold",
		"Selkirk Rex",
		"Siamese Cat",
		"Siberian",
		"Singapura",
		"Snowshoe",
		"Somali",
		"Sphynx",
		"Tonkinese",
		"Turkish Angora",
		"Turkish Van"
	];

	return breeds[Math.floor(Math.random()*breeds.length)];
}

function generateCats(amount)
{
	var cats = [];

	for (var i = 0; i < amount; i++)
	{
		cats.push(
			{
				kitty_id: i,
				name: randomName(),
				breed: randomBreed(),
				image: '/assets/catholder.png',
				generation: randomGeneration(),
				price: 1.0
			}
		);
	}
	return cats;
}

function successResponse(res, data)
{
	var ret = {
		data: data,
		api: 'faker'
	};
	res.send(ret);
	
}
app.listen(3000, () => console.log('Fake API listening on port 3000!'));