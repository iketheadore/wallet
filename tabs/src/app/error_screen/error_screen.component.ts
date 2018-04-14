import { Component, Input, OnInit} from '@angular/core';
import { finalize } from 'rxjs/operators';
import { ErrorScreenService } from './error_screen.service';

@Component({
  selector: 'error-screen',
  templateUrl: './error_screen.component.html',
  styleUrls: ['./error_screen.component.scss']
})

export class ErrorScreenComponent implements OnInit {
  currentError: any;
  constructor(private errorScreenService: ErrorScreenService) { 
  	
  }

  resetError() {
    this.errorScreenService.clearError();
  }
  ngOnInit() {
    this.errorScreenService.currentError.subscribe(error => {
      this.currentError = error;
    });
  }
}

