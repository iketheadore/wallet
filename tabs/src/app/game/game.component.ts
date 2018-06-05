import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer2, Input } from '@angular/core';
import { ScoreboardService } from './scoreboard.service';
import { finalize } from 'rxjs/operators';
import { environment } from '../../environments/environment';

@Component({
  selector: 'game',
  templateUrl: './game.component.html',
  styleUrls: ['./game.component.scss']
})
export class GameComponent implements OnInit {
 
  gameUrl: string;
  scores: any;
  span: string;
  isLoading: boolean;
  iframeHeight: number = 0;

  constructor(private scoreboardService: ScoreboardService, private renderer: Renderer2) { 
  	this.gameUrl = environment.serverUrl + "/scoreboard/game";
  }

 
  ngOnInit() {
    this.loadScores('day');

    let __this = this;
     // Listen to messages from parent window
    this.bindEvent(window, 'message', function (e:any) {
      if (e && e.data && e.data.command && e.data.command == "update height")
      {
          if (e.data.height && e.data.height > 0)
          {
            __this.iframeHeight = e.data.height;
          }
      }
    });
  }

   private bindEvent(element:any, eventName:any, eventHandler:any) {
        if (element.addEventListener) {
            element.addEventListener(eventName, eventHandler, false);
        } else if (element.attachEvent) {
            element.attachEvent('on' + eventName, eventHandler);
        }
    }


  loadScores(span: string) {
  	this.span = span;
  	this.isLoading = true;
    this.scoreboardService.getScores({span: this.span})
      .pipe(finalize(() => { this.isLoading = false; }))
      .subscribe((scores: any) => { this.scores = scores; });
  }
}
