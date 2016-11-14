package diff


// SliceDiff takes two slices and returns a slice which contains entries that were present in both.
func SliceDiff(slice1 []string, slice2 []string) ([]string){

	var diff []string

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}