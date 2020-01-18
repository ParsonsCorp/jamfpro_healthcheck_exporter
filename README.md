# JamfPro Health Check Exporter for Prometheus

The Jamf Pro health check page allows you to view the status of your environment.

This exporter is used to monitor that endpoint, and turn it into a scrapable metric.

| status | description |
| ------ | ----------- |
| [{"healthCode":1,"httpCode":503,"description":"DBConnectionError"}]       | An error occurred while testing the database connection. |
| [{"healthCode":2,"httpCode":200,"description":"SetupAssistant"}]          | The Jamf Pro Setup Assistant was detected. |
| [{"healthCode":3,"httpCode":503,"description":"DBConnectionConfigError"}] | A configuration error occurred while attempting to connect to the database. |
| [{"healthCode":4,"httpCode":503,"description":"Initializing"}]            | The Jamf Pro web app is initializing. |
| [{"healthCode":5,"httpCode":503,"description":"ChildNodeStartUpError"}]   | An instance of the Jamf Pro web app in a clustered environment failed to start. |
| [{"healthCode":6,"httpCode":503,"description":"InitializationError"}]     | A fatal error occurred and prevented the Jamf Pro web app from starting. |
| [ ]                                                                       | The Jamf Pro web app is running without error. |

Reference: [https://docs.jamf.com/10.18.0/jamf-pro/administrator-guide/Jamf_Pro_Health_Check_Page.html](https://docs.jamf.com/10.18.0/jamf-pro/administrator-guide/Jamf_Pro_Health_Check_Page.html)

## Docker Build Example

`docker build . -t jamfpro_healthcheck_exporter`

## Docker Run Example

List Help: `docker run -it --rm jamfpro_healthcheck_exporter -help`

Simple run: `docker run -it --rm -p 9613:9613 jamfpro_healthcheck_exporter -jamf.url="jamfpro.domain.com"`

Run with difference port: `docker run -it --rm -p 6060:6060 jamfpro_healthcheck_exporter -jamf.url="jamfpro.domain.com" -listen.port=6060`

Run with debug and color logrus: `docker run -it --rm -p 9613:9613 jamfpro_healthcheck_exporter -jamf.url="jamfpro.domain.com" -debug -enable-color-logs`

Run as daemon, with specific name, connected to jamf network, designate log driver, look to jamfpro on defined network, adjust protocal jamfpro is running on and debug: `docker run -d --rm --name=jamfpro_healthcheck_exporter --network=jamf --log-driver=json-file -p 9613:9613 jamfpro_healthcheck_exporter -jamf.url="jamfpro:8080" -jamf.proto=http -debug`

## References

Thank you everyone that writes code and docs!

- [https://golang.org/](https://golang.org/)
- [https://rsmitty.github.io/Prometheus-Exporters/](https://rsmitty.github.io/Prometheus-Exporters/)
- [https://prometheus.io/](https://prometheus.io/)
- [https://github.com/Sirupsen/logrus](https://github.com/Sirupsen/logrus)
- [https://www.jamf.com/](https://www.jamf.com/)
