# bt

## References and progress

Here are the specs I'm interested in exploring, along with their current implementation statuses:

* [BEP 3: The BitTorrent Protocol Specification](https://www.bittorrent.org/beps/bep_0003.html)
    * [x] bencoding
        * [x] parse Integer
        * [x] parse String
        * [x] parse List
        * [x] parse Dictionary
    * [ ] metainfo/.torrent implementation
        * [x] Unmarshal to Go struct
        * [ ] Marshal from Go struct
    * [ ] Tracker requests, response parsing
        * [ ] Unmarshal to Go struct
        * [ ] Marshal from Go struct
    * [ ] Peer protocol
* [BEP 4: Assigned Numbers](https://www.bittorrent.org/beps/bep_0004.html)
    * We'll want these as enums
* [BEP 5: DHT Protocol](https://www.bittorrent.org/beps/bep_0005.html)
    * Finding stuff
* [BEP 20: Peer ID Conventions](https://www.bittorrent.org/beps/bep_0020.html)
    * Identifying ourselves
* [BEP 29: uTorrent transport protocol (uTP)](https://www.bittorrent.org/beps/bep_0029.html)
    * ...maybe.
* [BEP 55: Holepunch extension](https://www.bittorrent.org/beps/bep_0055.html)
    * Getting past NAT
