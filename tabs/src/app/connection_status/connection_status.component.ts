import { Component, Input, OnInit} from '@angular/core';
import { finalize } from 'rxjs/operators';
import { ConnectionStatusService } from './connection_status.service';

@Component({
  selector: 'connection-status',
  templateUrl: './connection_status.component.html',
  styleUrls: ['./connection_status.component.scss']
})

export class ConnectionStatusComponent implements OnInit {

  status: any;

  constructor(private connectionStatusService: ConnectionStatusService) { }

  ngOnInit() {
  	this.connectionStatusService.status.subscribe(status => {
  		this.status = status;
  	});
  }

  doReconnect() {
  	this.connectionStatusService.doReconnect();
  }
}

