import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer2, Input } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { HttpClient } from '@angular/common/http';
import { MatDialogRef } from '@angular/material';

@Component({
  selector: 'scratchcard-dialog',
  templateUrl: './scratchcard_dialog.component.html',
  styleUrls: ['./scratchcard_dialog.component.scss']
})
export class ScratchCardDialogComponent implements OnInit {

  
  constructor(
    public dialogRef:MatDialogRef<ScratchCardDialogComponent>
  ) { }


  ngOnInit() {
   
  }

  close() {
    this.dialogRef.close();
  }
}
