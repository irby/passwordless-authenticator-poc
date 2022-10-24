export enum UserId {
    mirby7Id = "3280a1a2-9417-4b10-a6e9-987eabdf63ec",
    gburdell27Id = "da8c3048-78ee-470e-a9fb-c41a9b84de86",
    buzzId = "5bc3a580-d922-42f3-9031-a4faf8faef5d"
}

const mirby7WebauthnTokenId = "X-Xjt3TuMNWo-D8YR5BjNOUnTRE";
const gburdell27WebauthnTokenId = "Y-Xjt3TuMNWo-D8YR5BjNOUnTRE";
const buzzWebauthnTokenId = "Z-Xjt3TuMNWo-D8YR5BjNOUnTRE";

const mirby7UserHandle = "MoChopQXSxCm6Zh-q99j7A";
const gburdell27UserHandle = "2owwSHjuRw6p-8Qam4Tehg";
const buzzUserHandle = "W8OlgNkiQvOQMaT6-PrvXQ"

export function GetUserNameFromId(id: string): string | null {
    switch (id) {
        case UserId.mirby7Id:
            return "mirby7@gatech.edu";
        case UserId.gburdell27Id:
            return "gburdell27@gatech.edu";
        case UserId.buzzId:
            return "buzz@gatech.edu";
        default:
            return null;
    }
}

export function GetWebauthnTokenIdFromId(id: string): string {
    switch (id) {
        case UserId.mirby7Id:
            return mirby7WebauthnTokenId;
        case UserId.gburdell27Id:
            return gburdell27WebauthnTokenId;
        case UserId.buzzId:
            return buzzWebauthnTokenId;
        default:
            throw new Error("that shouldn't have happened");
    }
}

export function GetWebauthnUserHandleFromId(id: string): string {
    switch (id) {
        case UserId.mirby7Id:
            return mirby7UserHandle;
        case UserId.gburdell27Id:
            return gburdell27UserHandle;
        case UserId.buzzId:
            return buzzUserHandle;
        default:
            throw new Error("that shouldn't have happened");
    }
}