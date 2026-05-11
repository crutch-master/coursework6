vendor:
	mkdir -p web/static/vendor
	curl https://cdn.jsdelivr.net/npm/htmx.org@2.0.10/dist/htmx.min.js -o web/static/vendor/htmx.min.js
	curl https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css -o web/static/vendor/pico.min.css

.PHONY=vendor
