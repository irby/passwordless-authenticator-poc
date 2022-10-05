/* eslint-disable */
import { Injectable } from '@angular/core';
import axios, { AxiosRequestConfig, AxiosError } from 'axios';
import { Router } from '@angular/router';
import { ErrorResponse } from '../models/error-response.interface';
import { Observable, of } from 'rxjs';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from 'src/environments/environment';
import { ServiceError } from '../models/service-error.interface';
import { ServiceData } from '../models/service-data.interface';
import { ServiceResponse } from '../models/service-response.interface';

@Injectable()
export abstract class BaseService {
    private router: Router;
    private client: HttpClient;

    constructor(router: Router,
        client: HttpClient) {
        this.client = client;
        this.router = router;
    }

    protected createFullyQualifiedUrl(endpoint: string): string {
        return `${environment.hankoApiUrl}/${endpoint}`;
    }

    private createError<T>(error: AxiosError): ServiceError {
        // build this out

        const retError = {
            type: 'error'
        } as ServiceError;

        const errorResponse = error?.response?.data as ErrorResponse;

        if (!!errorResponse.errorCode) {
            retError.response = errorResponse;
        } else {
            retError.response = {} as ErrorResponse;
        }

        if (!!error.response) {
            retError.statusCode = error.response.status;
            retError.message = error.response.statusText;

            // if(error.response.data?.errorMessage){
            //     retError.body = error.response.data.errorMessage;
            // } else {
            //     retError.body = error.response.data.toString();
            // }

            // if (retError.statusCode === 401) {
            //     AuthenticationService.clearAuthentication();
            //     AuthenticationService.loginRedirect(this.router);
            // } else if (retError.statusCode === 403) {
            //     // do something
            // }
        }

        return retError;
    }

    private createData<T>(data: T): ServiceData<T> {
        console.log(typeof(data));
        return {
            type: 'data',
            data
        } as ServiceData<T>;
    }

    /*--- Post ---*/

    protected async postAsync<T>(
        endpoint: string,
        body?: {}
    ): Promise<ServiceResponse<T>> {
        let retResp = {} as ServiceResponse<T>;

        try {
            const resp = await axios.post<T>(
                this.createFullyQualifiedUrl(endpoint),
                body,
                this.createRequestConfig()
            );
            retResp = this.createData(resp.data);
        } catch (e) {
            retResp = this.createError<T>(e as AxiosError<T>);
        }

        return retResp;
    }

    /*--- Get ---*/
    protected getAsyncObservable<T>(endpoint: string): Observable<T> {
        let retResp = {} as ServiceResponse<T>;

        try {
            const resp = this.client.get<T>(this.createFullyQualifiedUrl(endpoint), { withCredentials: true });
            return resp;
        } catch (e) {
            // @ts-ignore
            return of(e);
        }
    }

    protected async getAsync<T>(endpoint: string): Promise<ServiceResponse<T>> {
        let retResp = {} as ServiceResponse<T>;

        try {
            const resp = await axios.get<T>(
                this.createFullyQualifiedUrl(endpoint),
                this.createRequestConfig()
            );
            retResp = this.createData(resp.data);
        } catch (e) {
            // @ts-ignore
            retResp = this.createError<T>(e);
        }

        return retResp;
    }

    /*---- Put ----*/

    protected async putAsync<T>(
        endpoint: string,
        body?: {}
    ): Promise<ServiceResponse<T>> {
        let retResp = {} as ServiceResponse<T>;

        try {
            const resp = await axios.put<T>(
                this.createFullyQualifiedUrl(endpoint),
                body,
                this.createRequestConfig()
            );
            retResp = this.createData(resp.data);
        } catch (e) {
            retResp = this.createError<T>(e as AxiosError<T>);
        }

        return retResp;
    }

    /*---- Delete ----*/

    protected async deleteAsync<T>(
        endpoint: string
    ): Promise<ServiceResponse<T>> {
        let retResp = {} as ServiceResponse<T>;

        try {
            const resp = await axios.delete<T>(
                this.createFullyQualifiedUrl(endpoint),
                this.createRequestConfig()
            );
            retResp = this.createData(resp.data);
        } catch (e) {
            retResp = this.createError<T>(e as AxiosError<T>);
        }

        return retResp;
    }

    /*---- Authentication ----*/

    private createRequestConfig<T>(): AxiosRequestConfig {
        const config = {} as AxiosRequestConfig;

        config.transformResponse = (data) => {
            return this.deserialize(data);
        };

        config.withCredentials = true;

        return config;
    }

    private deserialize<T>(data: string): T {
        return JSON.parse(data, this.reviveDateTime) as T;
    }

    private reviveDateTime(key: any, value: any): any {
        if (typeof value === 'string') {
            const a = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)(?:([\+-])(\d{2})\:(\d{2}))?Z?$/.exec(value);
            if (a) {
                const utcMilliseconds = Date.UTC(+a[1], +a[2] - 1, +a[3], +a[4], +a[5], +a[6]);
                return new Date(utcMilliseconds);
            }
        }
        return value;
    }
}
