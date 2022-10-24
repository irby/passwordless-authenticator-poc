import { Injectable } from "@angular/core";
import { ToastrService, IndividualConfig } from 'ngx-toastr';

@Injectable({ providedIn: 'root' })
export class NotificationService {
    private readonly errorConfig: Partial<IndividualConfig> = {
        disableTimeOut: true,
        closeButton: true,
    };

    private readonly dismissibleErrorConfig: Partial<IndividualConfig> = {
        disableTimeOut: false,
        timeOut: 5000,
        progressBar: true,
        closeButton: true
    };

    private readonly infoConfig: Partial<IndividualConfig> = {
        disableTimeOut: false,
        timeOut: 5000,
        progressBar: true,
        closeButton: true
    };

    private readonly warningConfig: Partial<IndividualConfig> = {
        disableTimeOut: true,
        closeButton: true
    };

    private readonly successConfig: Partial<IndividualConfig> = {
        disableTimeOut: false,
        timeOut: 3000,
        progressBar: true,
        closeButton: true,
    };

    private readonly disableTimeoutConfig: Partial<IndividualConfig> = {
        disableTimeOut: true,
        progressBar: false,
        closeButton: true,
    };

    constructor(private readonly toast: ToastrService) {
    }

    public error(message: string, title?: string): void {
        title = title || 'Error';
        this.toast.error(message, title, this.errorConfig);
    }

    public dismissibleError(message: string, title?: string): void {
        title = title || 'Error';
        this.toast.error(message, title, this.dismissibleErrorConfig);
    }

    public info(message: string, autoDismiss: boolean, title?: string): void {
        title = title || 'Info';
        const config = this.infoConfig;
        config.disableTimeOut = !autoDismiss;
        config.progressBar = autoDismiss;

        this.toast.info(message, title, this.infoConfig);
    }

    public warning(message: string, title?: string): void {
        title = title || 'Warning';
        this.toast.warning(message, title, this.dismissibleErrorConfig);
    }

    public success(message: string, title?: string): void {
        title = title || 'Success';
        this.toast.success(message, title, this.successConfig);
    }

    public successStatic(message: string, title?: string): void {
        title = title || 'Success';
        this.toast.success(message, title, this.disableTimeoutConfig);
    }
}