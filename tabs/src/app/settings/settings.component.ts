import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer, Input, Renderer2 } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { MatDialogRef } from '@angular/material';
import { SettingsService } from './settings.service';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {
 
  restore_name: string;
  restore_seed: string;

  constructor(private dialogRef:MatDialogRef<SettingsComponent>, 
              private renderer: Renderer2,
              private settingsService: SettingsService) { }

 
  ngOnInit() {
  }

  doClose() {
    this.dialogRef.close();
  }

  doBackup() {
    alert("Backup isn't functioning yet");
  }

  doRestore() {
    let params = {
      label: this.restore_name,
      seed: this.restore_seed,
      aCount: 1,
      encrypted: false
    };
    this.settingsService.restoreSeed(params).subscribe((result: any) => { 
      let refresh_event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
      document.dispatchEvent(refresh_event);
      this.dialogRef.close();
    });
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
