import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import { environment } from '../../environments/environment';

const routes = {
  all_statuses: () => `${environment.walletUrl}/conn/all_statuses`,
  status: () => `${environment.walletUrl}/conn/status`,
  reconnect: () => `${environment.walletUrl}/conn/reconnect`,
};

@Injectable()
export class ConnectionStatusService {

  private all_statuses: any = false;

  private statusSource = new BehaviorSubject<any>(false);
  status = this.statusSource.asObservable();

  constructor(private httpClient: HttpClient) { 
    this.getAllStatuses().subscribe((statuses: any) => { 
      this.all_statuses = statuses;
        this.checkStatus();
    });
  }

  private checkStatus()
  {
    //Now get the status
    this.getStatus().subscribe((status_code: number) => {
      this.statusSource.next(this.lookupStatus(status_code));
    });
  }

  doReconnect()
  {
    this.reconnect().subscribe((res: any) => {
      this.checkStatus();
      return;
    });
  }

  private lookupStatus(status_code: number)
  {
    let status = this.all_statuses.find(status => status.code === status_code);
    let __this = this;
    clearInterval(status.timer);
    if (status.code === 1)
    {
      //Set a timer
      status.retry_time = 30;
      status.timer = setInterval(() => { 
        status.retry_time = status.retry_time - 1;
        if (status.retry_time <= 0)
        {
          clearInterval(status.timer);
          __this.doReconnect();
          return;
        }
      }, 1000);
    }
    return status;
  }

  private reconnect(): Observable<any> {
    return this.httpClient
      .get(routes.reconnect())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not reconnect :-('))
      );
  }

  private getStatus(): Observable<any> {
    return this.httpClient
      .get(routes.status())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not load status :-('))
      );
  }

  private getAllStatuses(): Observable<any> {
    return this.httpClient
      .get(routes.all_statuses())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not load all_statuses :-('))
      );
  }

}
