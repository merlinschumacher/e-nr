# e-nr.de
## E-Nummern und Inhaltsstoffe einfach abfragen.

Mit dieser Seite können Sie einfach die Bedeutung von <a
    href="https://de.wikipedia.org/wiki/Lebensmittelzusatzstoff" target="_blank">E-Nummern</a> in den
Inhaltsstoffen von Produkten ermitteln.
Besuchen Sie zum Beispiel [300.e-nr.de](https://300.e-nr.de) um zu erfahren, was
sich
hinter E300 verbirgt. Alternativ können Sie auch [kaliumacetat.e-nr.de](https://kaliumacetat.e-nr.de)
besuchen um Anhand des Namens mehr über den Inhaltsstoff zu erfahren.
Alle Links führen direkt zum entsprechenden Artikel der deutschsprachigen Wikipedia.

### Spaß für Nerds mit DNS
Für technikinteressierte Personen stellt diese Seite auch einen DNS-Dienst bereit über den man die Bedeutung
der E-Nummern direkt abrufen kann. Mit dem Kommando dig lassen sich Einträge wie folgt abfragen:

    dig 100.e-nr.de @e-nr.de

Der Dienst liefert A-, AAAA-, CNAME-, URI- und TXT-Records aus. A und AAAA geben lediglich die IP des
Servers und die alternativen Namen der E-Nummer aus. CNAME gibt ebenfalls die alternativen Namen aus. Der
URI-Eintrag gibt den Link zum Wikipedia-Artikel aus. Der TXT-Eintrag enthält den vollen Namen, die
Beschreibung und die Wikipedia-URL.

Der Dienst ist in Go geschrieben und verwendet hauptsächlich die [DNS Bibliothek von Miekg](https://github.com/miekg/dns).