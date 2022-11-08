export enum UserId {
  mirby7Id = "3280a1a2-9417-4b10-a6e9-987eabdf63ec",
  gburdell27Id = "da8c3048-78ee-470e-a9fb-c41a9b84de86",
  buzzId = "5bc3a580-d922-42f3-9031-a4faf8faef5d",
  adminId = "4280a1a2-9417-4b10-a6e9-087eabdf63ed",
}

export function GetUserNameFromId(id: string): string | null {
  switch (id) {
    case UserId.mirby7Id:
      return "mirby7@gatech.edu";
    case UserId.gburdell27Id:
      return "gburdell27@gatech.edu";
    case UserId.buzzId:
      return "buzz@gatech.edu";
    case UserId.adminId:
      return "admin@gatech.edu";
    default:
      return null;
  }
}
