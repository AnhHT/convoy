import { Injectable } from '@angular/core';
import { HTTP_RESPONSE } from 'src/app/models/http.model';
import { HttpService } from 'src/app/services/http/http.service';

@Injectable({
	providedIn: 'root'
})
export class CreateGroupService {
	constructor(private http: HttpService) {}

	async createGroup(
		requestDetails: { name: string; retry: { interval_seconds: string; limit: string }; signature: { header: string; hash: string; disable_endpoint: boolean } },
		requestOptions: { orgId: string }
	): Promise<HTTP_RESPONSE> {
		try {
			const response = await this.http.request({
				url: `groups?${requestOptions.orgId || ''}`,
				body: requestDetails,
				method: 'post'
			});
			return response;
		} catch (error: any) {
			return error;
		}
	}

	async editGroup(
		requestDetails: { name: string; retry: { interval_seconds: string; limit: string }; signature: { header: string; hash: string; disable_endpoint: boolean } },
		requestOptions: { orgId: string; groupId: string }
	): Promise<HTTP_RESPONSE> {
		try {
			const response = await this.http.request({
				url: `groups/${requestOptions.groupId}?${requestOptions.orgId || ''}`,
				body: requestDetails,
				method: 'put'
			});
			return response;
		} catch (error: any) {
			return error;
		}
	}
}
