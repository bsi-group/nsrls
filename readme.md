nsrls (NSRL Server)
===================

nsrls is a server designed to provide access to the NSRL hash data set. There are 
two methods to access the data either running as a single process or using a JSON 
HTTP API. The HTTP JSON API can be used by the nsrlc (NSRL Client) or directly by 
other applications or processes.

## Importing ##

The **nsrls** application can perform show data manipulation when importing the 
data set. It can extract a specific field from a CSV file by using the **-s** or 
**--csvfield** parameters, so if the second field is to be used the use the value 
**2** for the parameter. If the import file has quotes around the hash data then 
they can be removed when importing by using the **-r** or **--removequotes** parameters. 

## HTTP API ##

The HTTP API has two different acess methods; a single hash can be checked using 
a HTTP GET request or a bulk request can be performed using a HTTP POST request.  

The IP local IP address and port that the HTTP server runs on are configured 
using the config file (nsrls.config). The **show_requests** option in the config 
file determines whether the HTTP requests against the server are displayed in the 
console. 

The HTTP API is located by default at the following URL's:

```
    http://127.0.0.1:8080/single (GET)
    http://127.0.0.1:8080/bulk   (POST)
```

### Single ###

The single API URL takes a hash value in the URL like so:

```
    http://127.0.0.1:8080/single/392126E756571EBF112CB1C1CDEDF926
```

### Bulk ###

The bulk API uses a HTTP POST request like so, with the hashes each separated 
by a hash (#) character:

```
    POST http://127.0.0.1:8080/bulk HTTP/1.1
    Content-Type: application/x-www-form-urlencoded
    Content-Length: 65
    
    392126E756571EBF112CB1C1CDEDF926#8E23576EF5AEF2D5457C8A24BF5F740A
```

The NSRL client application by default sends up to 1000 hashes per batch.

### Return Data ###

The HTTP API returns data in the JSON format as shown below:

```
    {"Hash":"8E23576EF5AEF2D5457C8A24BF5F740A","Exists":true}
```

## Single Use (File) Mode ##

The server can be used to perform a single use lookup against a input file using
the **-m** parameter and a value of **f** (file). The server will import the hash
data, process the input file containing multiple hashes and extract the data to 
an output file.

### Output ###

When in single use mode, the application outputs directly to a file. The output 
format can be defined by the command line parameters. The options for the output
format are:
 
- i: Outputs only the identified hashes
- u: Outputs only the unidentified hashes
- a: Outputs both the identified and unidentified hashes, along with a status column

## Configuration ##

- Make sure all of the paths specified in the config file are fully qualified
- Ensure that there is a log directory created, and that the user can write to it
```
     sudo mkdir /var/log/nsrls
     sudo chown <username> /var/log/nsrls
```