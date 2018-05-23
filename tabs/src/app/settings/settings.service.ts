import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { environment } from '../../environments/environment';

const routes = {
  new_wallet: (s: WalletContext) => `http://127.0.0.1:6148/api/wallets/new`
};

export interface WalletContext {
  label: string;
  seed: string;
  aCount: number;
  encrypted: boolean;
}

@Injectable()
export class SettingsService {

  constructor(private httpClient: HttpClient) { }

  restoreSeed(context: WalletContext): Observable<object> {
    return this.httpClient
      .post(routes.new_wallet(context), this.getQueryString(context), this.getOptions())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not restore seed :-('))
      );
  }

  private getQueryString(parameters:any = null) {
    if (!parameters) {
      return '';
    }

    return Object.keys(parameters).reduce((array,key) => {
      array.push(key + '=' + encodeURIComponent(parameters[key]));
      return array;
    }, []).join('&');
  }


  private getOptions() {
    const headers = {
      'Content-Type': 'application/x-www-form-urlencoded',
    };

    return { headers: headers };
  }
}
