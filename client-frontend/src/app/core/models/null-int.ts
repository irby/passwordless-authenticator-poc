export class NullInt {
    private readonly int32: number;
    private readonly valid: false;

    constructor(int32: number, valid: false) {
        this.int32 = int32;
        this.valid = valid;
    }
    public getValueOrDefault() : number | null {
        console.log(this.valid, this.int32);
        return !this.valid ? null : this.int32;
    }
}

export interface NullInt32 {
    Int32: number;
    Valid: false;
}