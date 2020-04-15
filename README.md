# REST API, database & image handling in go

REST API used to help populate sqlite database with attractions in Lithuania.
The program also handles images from an url.Used with [frontend](https://github.com/MingaudasVagonis/vue-attractions-front).

 - [**Retrieving data & images**](#retrieving-data-and-images)  
 - [**API**](#api)  
 - [**Attractions' & database structure**](#attractions-and-database-structure)  
 - [**Libraries & usage**](#libraries-and-usage)  
 
## Retrieving data and images

**see** [**retrieve.go**](retrieve.go)

***merge** [target database url] [optional: url used to post the images]*

Merging process consists of the following steps:

 - Adding data from cache database to the target database
 - Downloading images
 - Processing images
 - Saving them locally or posting them to the url provided
 <img src="https://i.imgur.com/LRkWx3T.png" height="300"/>
 
## API

**see** [**server.go**](server.go)

API has the following two routes

 ### *add* [POST]
 **Used to add an attraction to the database**
 Requst body must contain a json object with fields:
 
 - **category** string **|** must be one of: nature, heritage, museums
 - **description** json object
	  - **name** string **|** must be longer than 3 
	  - **hours** json object 
		  - **wkd** string **|** must match *([0-9]{2}:[0-9]{2}-[0-9]{2}:[0-9]{2})* pattern
		  - **std** string **|** must match *([0-9]{2}:[0-9]{2}-[0-9]{2}:[0-9]{2})* pattern
		  - **snd** string **|** must match *([0-9]{2}:[0-9]{2}-[0-9]{2}:[0-9]{2})* pattern
	- **info** string **|** must be longer than 30 characters
- **location** json object
  - **city** string **|** must be longet than 3 characters and contain only lithuanian alphabet
  - **coordinates** json object
    - **latitude** number **|** must fall between 53.53 and 56.27
    - **longitude** number  **|** must fall between 20.56 and 26.5
- **image** json object
  - **url** string **|** may be null
  - **copyright** string **|** may be null

 ### *check* [GET]
 **Used to check whether the attraction allready exists in the database.**
Request must contain the following query paramters:

 - **name** string | name of the object

Otherwise response status will be 404.

 
## Attractions' and database structure

**see** [**db.go**](db.go) [**attraction.go**](attraction.go)

Cache stores attraction objects in the following columns

 - **id** text, not null
 - **category** text, not null
 - **description** text, not null
	 - description is a stringified json object that consits of:
		 - **name** string
		 - **hours** json object
		 - **info** string
 - **location** text, not null
	 - location is a stringified json object that consits of:
		 - **city** string
		 - **coordinates** json object
 - **copyright** text
 - **url** text

*Target database schema does not contain url column*

Cache stores data used to check whether an attraction already exists in the following columns

- **compare** string **|** value used to compare the names a.k.a id
- **display** string **|** value used to display results to the user

## Libraries and usage

### Following libraries are used
 - [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
 - [github.com/gorilla/mux](https://github.com/gorilla/mux)
 - [github.com/nfnt/resize](https://github.com/nfnt/resize)
 - [github.com/oliamb/cutter](https://github.com/oliamb/cutter)
 
### One time launch: 
```
  git clone https://github.com/MingaudasVagonis/go-attractions-server.git
  cd go-attractions-server
  go run main.go db.go attraction.go server.go utils.go retrieve.go
```

### Commands
  ***merge** [target database url] [optional: url used to post the images]*
  - [Retrieving data & images](#retrieving-data-and-images) 
      
  ***initialize** [external database url]*
  - Adds data used to check whether the attraction exists from an external database.
