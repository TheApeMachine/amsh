using Go = import "/go.capnp";
@0xe363a5839bf866c4;
$Go.package("data");
$Go.import("data/artifact");

struct Artifact {
  id @0 :Text;
  checksum @1 :Data;
  pubkey @2 :Data;
  version @3 :Text;
  type @4 :Text;
  timestamp @5 :UInt64;
  origin @6 :Text;
  role @7 :Text;
  scope @8 :Text;
  attributes @9 :List(Attribute);
  payload @10 :Data;
}

struct Attribute {
  key @0 :Text;
  value @1 :Text;
}

interface ModelService {
  query @0 (request :Artifact) -> (response :Artifact);
}