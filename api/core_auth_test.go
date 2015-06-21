package api

import "testing"

func TestHashPassowrd(t *testing.T) {

	type Passwords struct {
		Original  string
		Encrypted string
	}

	tests := map[Passwords]bool{
		Passwords{"qweww", "278924da841f2cd2c494a5f39b108836d75d6ef0aea0cec7aa90a9a85a90a7ace23386e0559b577f"}:               true,
		Passwords{"thegoodsareinthesky", "ca5906835b8093baf5555b9f5a3d36227e0bc241b44dd646561d490fd7d3e2fcbd834ac3c96bfc6f"}: true,
		Passwords{"thegoodsareinthesky", "fb6c1037cbfba8a6c1bb86536b47b610a716effdb9a89c8d539bc0e24c298f90678fcf4b61738a0f"}: true,
		Passwords{"cup", "f448c50b601715bdd6ea82697eee179cbe868eae5d61bdca4f05307590045dbd6bfd00b4c5ac5a5a"}:                 true,
		Passwords{"cup", "a6b1a073df090c96f7f2bc6b378f5fbff733c06566a82eb66a495402dabdb20448d1bb8fbac3518c"}:                 true,
		Passwords{"lampara", "7695107cda051605aa7d4f24b63d70e444674d5c810fd5d2ed5e4dadd7219da4fe8f4bb12ccab582"}:             true,
		Passwords{"lampara", "0739f17cefeaa5c98bc1b6f754b94f9a0e2e1360d572fa402fd6190b66d226a8e4fe628088b98f9a"}:             true,
		Passwords{"lampara", "12359af317b844d87b7864f8cb1bd750773a7a163ca7027274c39595ace5e147c3242e499079c326"}:             false,
		Passwords{"Santo Niño de Cebú", "2fba7e078a17f87c69df4288d0878cedf53baaee5b3295022f147afecf83b84376b5c28b54622767"}:  true,
		Passwords{"Santo Niño de Cebú", "5e17359c5e55ce31dd663a5abf374a4edf63170090e387e6689ebe9076ed8ea73266c61d0111b2ad"}:  true,
		Passwords{"Santo Niño de Cebú", "1234c492489b9663b1a5a5e8d8b3d5236686738ab4cfdf890e9f8b34bb65788e826edc5776fb2af8"}:  false,
		Passwords{"большинства", "696fe0c215062b8c35ea150d7179bf518633ad9fd1808e6eb000ae7c819ada841b53325a15361d16"}:         true,
		Passwords{"большинства", "d71ecdf956d61300ec56aa47ef1caaf2e628807203639e4931beb9f0f66103fbe9f976f427bea93c"}:         true,
		Passwords{"большинства", "c6e9af267843bc01ccfa76b959d502da6c8e84698fc7c675bed6c528b2648b4b6e0a469dba79fa77"}:         false,
	}

	for i, ok := range tests {
		if CheckPassword(i.Original, i.Encrypted) != ok {
			t.Errorf("Expected %s to be %b", i.Original, ok)
		}
		// tests[i]
	}
}

/*
func TestConnectDB(t *testing.T) {
	db := db.SetupDB()

	err := db.Database.Ping()
	if err != nil {
		t.Fail("Could not connect to database %v", err)
	}
}
*/
