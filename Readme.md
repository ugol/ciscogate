# Ciscogate

Bla bla 

## Usage

Building and compiling on OpenShift

oc new-app https://github.com/ugol/ciscogate.git

Available Environents:
 CISCO_ADDRESS to set listen interface/webhook. Default is 'localhost:8080'
 CISCO_APICURL to set apic url. Default is 'apic1.rmlab.local' 
 CISCO_APICUSERNAME to set user name. Default is 'admin' 
 CISCO_APICPASSWORD to set password. Default is 'C1sco123' 
 CISCO_OPENSHIFTTENANT to set OCP tenant. Default is 'openshift39' 
 CISCO_EPGTOBECREATED to set epg. Default is 'prova18e26' 
 CISCO_GRACEFULTIMEOUT to set GracefulTimeout. Default is '15s' 
 CISCO_WRITETIMEOUT to set WriteTimeout. Default is '15s' 
 CISCO_READTIMEOUT to set ReadTimeout. Default is '15s'
 CISCO_IDLETIMEOUT to set IdleTimeout. Default is '1m0s'


```bash
cd ciscogate
go build
go get -u
```

## Usage

...
