package utils

func Must(err error) {
    if err != nil {
        panic(err)
    }
}

func Must1[T any](val T, err error) T {
    if err != nil {
        panic(err)
    }
    return val
}

func Must2[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
    if err != nil {
        panic(err)
    }
    return val1, val2
}

