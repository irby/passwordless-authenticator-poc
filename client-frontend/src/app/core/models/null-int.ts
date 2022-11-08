export class NullInt {
  private readonly int32: number;
  private readonly valid: boolean;

  constructor(int32: number, valid: boolean) {
    this.int32 = int32;
    this.valid = valid;
  }

  public getValueOrDefault(): number | null {
    return !this.valid ? null : this.int32;
  }
}

export interface NullInt32 {
  Int32: number;
  Valid: boolean;
}
