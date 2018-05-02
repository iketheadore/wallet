import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer, Input, Renderer2 } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { MatDialogRef } from '@angular/material';

@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {
 
 
  constructor(private dialogRef:MatDialogRef<SettingsComponent>, private renderer: Renderer2) { }

 
  ngOnInit() {
  }

  doClose() {
    this.dialogRef.close();
  }

  doBackup() {
    alert("Backup isn't functioning yet");
  }

  doRestore() {
    alert("Restore isn't functioning yet");
  }

  toggleFrame() {
    if (document.body.classList.contains('colored-frame'))
    {
      this.renderer.removeClass(document.body, 'colored-frame');
    }
    else
    {
      this.renderer.addClass(document.body, 'colored-frame');
    }
    
  }

}
