import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { environment } from '../../environments/environment';

const routes = {
  new_wallet: (s: WalletContext) => `http://127.0.0.1:6148/v1/wallets/new`,
  list_wallets: () => `http://127.0.0.1:6148/v1/wallets/list`,
  get_wallet: () => `http://127.0.0.1:6148/v1/wallets/get`
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

  getBackupFile(): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      this.getWalletList().then(wallets => {

        let promises = [];

        for (let i = 0; i < wallets.length; i++)
        {
          let wallet = wallets[i];
          promises.push(this.getWalletDetails(wallet.label));
        }
        Promise.all(promises).then(wallets=> {
          resolve(wallets);
        })
      });
    });
  }

  private getWalletList(): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      this.httpClient
      .get(routes.list_wallets(), this.getOptions()).subscribe((wallets:any) => {

        if (wallets && wallets.wallets)
        {
          resolve(wallets.wallets);
        }
        else
        {
          reject("No wallets found");
        }
        
      });
    });
  }

  private getWalletDetails(label: string): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      this.httpClient
      .post(routes.get_wallet(), this.getQueryString({label: label}), this.getOptions()).subscribe((wallet:any) => {
        if (wallet && wallet.entry_count && wallet.entry_count > 0)
        {
          resolve(wallet);
        }        
        else
        {
          reject("Wallet not found");
        }
      });
    });
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
