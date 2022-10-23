class Keys {
    static mirby7PrivateKey: string = "d080957316cac94b04d216377210de3f5b23bc0e10b701bc3c41db9bc4b48ac0";
    static mirby7XValue: string = "d6b5d57af9afef9f3fb31ccf532d25338570c86f48209b43d8119a55358c8b5c";
    static mirby7YValue: string = "27f0b0210f21c9f6558adc6ef07653bc04a3bcc1c3a4fadf429c2f075c499229";

    static gburdell27PrivateKey: string = "d276448670da90729cd73b60251e23f153f7d7eaa1ad2156a5b4b0a51f13ff8e";
    static gburdell27XValue: string = "d54f5cbd187b02a8f991d99f38542a31d78b9e2ccec94f1d4660e21272cd4b59";
    static gburdell27YValue: string = "23eaf9352a5b131b077057c7bed102d97c5343a0b1b8607d8ab7116c1dc89a70";

    static buzzPrivateKey: string = "8ab1bdc388ee1e767e03c6667222332dbb8e76624df7e712d2a8a0bd90455149";
    static buzzXValue: string = "440a1aeec3d6b2f9d29c6cb32460b6675b9679962eeb137791da0d7fbb9b9b4c";
    static buzzYValue: string = "651807e3b357c9697e18c4664fced1242f54fabc8588848f9934c36215adf429";
}

export function getPrivateKeyByName(name: string): string {
    switch(name) {
        case "mirby7@gatech.edu":
            return Keys.mirby7PrivateKey;
        case "gburdell27@gatech.edu":
            return Keys.gburdell27PrivateKey;
        case "buzz@gatech.edu":
            return Keys.buzzPrivateKey;
        default:
            throw new Error("user not found");
    }
}